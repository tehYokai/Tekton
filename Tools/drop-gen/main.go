package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/projectdiscovery/goflags"
	"golang.org/x/crypto/ssh"
)

type Droplet struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Region struct {
		Slug string `json:"slug"`
	} `json:"region"`
	Size  string `json:"size"`
	Image struct {
		Slug string `json:"slug"`
	} `json:"image"`
	Networks struct {
		V4 []struct {
			IPAddress string `json:"ip_address"`
		} `json:"v4"`
	} `json:"networks"`
}

type CreateDroplet struct {
	Name    string   `json:"name"`
	Region  string   `json:"region"`
	Size    string   `json:"size"`
	Image   string   `json:"image"`
	SSHKeys []string `json:"ssh_keys"`
}

type CreateDropletResponse struct {
	Droplet struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Region struct {
			Slug string `json:"slug"`
		} `json:"region"`
		Size struct {
			Slug string `json:"slug"`
		} `json:"size"`
		Image struct {
			Slug string `json:"slug"`
		} `json:"image"`
	} `json:"droplet"`
}

type Droplets struct {
	Droplets []Droplet `json:"droplets"`
}

type SSHKey struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

type Size struct {
	Slug         string   `json:"slug"`
	Memory       int      `json:"memory"`
	Vcpus        int      `json:"vcpus"`
	Disk         int      `json:"disk"`
	Transfer     int      `json:"transfer"`
	PriceMonthly float64  `json:"price_monthly"`
	PriceHourly  float64  `json:"price_hourly"`
	Regions      []string `json:"regions"`
}

type Sizes struct {
	Sizes []Size `json:"sizes"`
}

type options struct {
	listDroplets bool
	deleteAll    bool
	fleet        string
	amount       int
	sizes        bool
	size         string
}

func listSizes() {
	req, err := http.NewRequest("GET", "https://api.digitalocean.com/v2/sizes", nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("DO_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var sizes Sizes
	json.Unmarshal(body, &sizes)

	for _, size := range sizes.Sizes {
		fmt.Printf("Slug: %s, Memory: %d MB, VCPUs: %d, Disk: %d GB, Transfer: %d TB, Price: $%.2f/month, $%.2f/hour\n",
			size.Slug, size.Memory, size.Vcpus, size.Disk, size.Transfer, size.PriceMonthly, size.PriceHourly)
	}
}

func listDroplets() {
	req, err := http.NewRequest("GET", "https://api.digitalocean.com/v2/droplets", nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("DO_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var droplets Droplets
	json.Unmarshal(body, &droplets)

	dropsJSON, err := ioutil.ReadFile("drops.json")
	if err != nil {
		log.Fatal(err)
	}

	var existingDroplets Droplets
	if err := json.Unmarshal(dropsJSON, &existingDroplets); err != nil {
		log.Fatal(err)
	}

	for i, droplet := range droplets.Droplets {
		ipAddress := ""
		if len(droplet.Networks.V4) > 0 {
			ipAddress = droplet.Networks.V4[0].IPAddress
		}

		if i < len(existingDroplets.Droplets) {
			existingDroplets.Droplets[i].Networks.V4 = append(existingDroplets.Droplets[i].Networks.V4, struct {
				IPAddress string `json:"ip_address"`
			}{IPAddress: ipAddress})
		}
	}

	dropsJSON, err = json.Marshal(existingDroplets)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("drops.json", dropsJSON, 0644)
	if err != nil {
		log.Fatal(err)
	}

	if len(droplets.Droplets) == 0 {
		fmt.Println("No droplets to display.")
		return
	}

	for _, droplet := range droplets.Droplets {
		ipAddress := ""
		if len(droplet.Networks.V4) > 0 {
			ipAddress = droplet.Networks.V4[0].IPAddress
		}
		fmt.Printf("ID: %d, Name: %s, Region: %s, Image: %s, IP: %s\n",
			droplet.ID, droplet.Name, droplet.Region.Slug, droplet.Image.Slug, ipAddress)
	}
}

func deleteDroplets() {
	req, err := http.NewRequest("GET", "https://api.digitalocean.com/v2/droplets", nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("DO_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var droplets Droplets
	json.Unmarshal(body, &droplets)

	if len(droplets.Droplets) == 0 {
		fmt.Println("No droplets to delete.")
		return
	}

	for _, droplet := range droplets.Droplets {
		deleteURL := fmt.Sprintf("https://api.digitalocean.com/v2/droplets/%d", droplet.ID)
		req, err := http.NewRequest("DELETE", deleteURL, nil)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+os.Getenv("DO_API_KEY"))

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			log.Printf("Failed to delete droplet with ID %d\n", droplet.ID)
		} else {
			log.Printf("Deleted droplet with ID %d\n", droplet.ID)
			deleteKeyDirectory(droplet.Name)
		}
	}
}

func deleteKeyDirectory(dropletName string) {
	keyDir := filepath.Join("keys", dropletName)
	err := os.RemoveAll(keyDir)
	if err != nil {
		log.Printf("Failed to delete key directory for droplet '%s': %s\n", dropletName, err)
	} // else {
	// log.Printf("Deleted key directory for droplet '%s'\n", dropletName)
	// }
}

// func contains(droplets []Droplet, droplet Droplet) bool {
// 	for _, d := range droplets {
// 		if d.ID == droplet.ID {
// 			return true
// 		}
// 	}
// 	return false
// }

func createDroplet(name string, count int) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privBLK := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	privPEM := pem.EncodeToMemory(&privBLK)
	if err := ioutil.WriteFile("id_rsa", privPEM, 0600); err != nil {
		log.Fatal(err)
	}

	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatal(err)
	}

	pubBytes := ssh.MarshalAuthorizedKey(pub)
	if err := ioutil.WriteFile("id_rsa.pub", pubBytes, 0644); err != nil {
		log.Fatal(err)
	}

	sshKey := SSHKey{
		Name:      name,
		PublicKey: string(pubBytes),
	}

	sshKeyJSON, err := json.Marshal(sshKey)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", "https://api.digitalocean.com/v2/account/keys", bytes.NewBuffer(sshKeyJSON))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("DO_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatal(err)
	}

	keyID := strconv.Itoa(int(result["ssh_key"].(map[string]interface{})["id"].(float64)))

	dropsJSON, err := ioutil.ReadFile("drops.json")
	if err != nil {
		log.Fatal(err)
	}

	existingDroplets := Droplets{}
	if err := json.Unmarshal(dropsJSON, &existingDroplets); err != nil {
		log.Fatal(err)
	}

	startCount := len(existingDroplets.Droplets) + 1

	droplets := make([]Droplet, 0)

	for i := startCount; i < startCount+count; i++ {
		dropletName := fmt.Sprintf("%s%d", name, i)

		droplet := CreateDroplet{
			Name:    dropletName,
			Region:  "nyc1",
			Size:    "s-1vcpu-1gb",
			Image:   "ubuntu-20-04-x64",
			SSHKeys: []string{keyID},
		}

		dropletJSON, err := json.Marshal(droplet)
		if err != nil {
			log.Fatal(err)
		}

		req, err := http.NewRequest("POST", "https://api.digitalocean.com/v2/droplets", bytes.NewBuffer(dropletJSON))
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+os.Getenv("DO_API_KEY"))

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		var dropletResponse CreateDropletResponse
		if err := json.Unmarshal(body, &dropletResponse); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Created droplet with ID: %d\n", dropletResponse.Droplet.ID)

		dropletURL := fmt.Sprintf("https://api.digitalocean.com/v2/droplets/%d", dropletResponse.Droplet.ID)
		req, err = http.NewRequest("GET", dropletURL, nil)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+os.Getenv("DO_API_KEY"))

		resp, err = client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		var dropletDetails Droplet
		if err := json.Unmarshal(body, &dropletDetails); err != nil {
			log.Fatal(err)
		}

		newDroplet := Droplet{
			ID:     dropletResponse.Droplet.ID,
			Name:   dropletResponse.Droplet.Name,
			Region: dropletResponse.Droplet.Region,
			Size:   dropletResponse.Droplet.Size.Slug,
			Image:  dropletResponse.Droplet.Image,
		}

		droplets = append(droplets, newDroplet)

		// Save SSH key for the droplet
		saveSSHKey(dropletName, privPEM, pubBytes)
	}

	existingDroplets.Droplets = append(existingDroplets.Droplets, droplets...)

	dropsJSON, err = json.Marshal(existingDroplets)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("drops.json", dropsJSON, 0644)
	if err != nil {
		log.Fatal(err)
	}

	// Wait for 10 seconds before listing droplets
	fmt.Println("Waiting for 1 minute to get droplet IPs")
	time.Sleep(60 * time.Second)

	// List droplets
	listDroplets()
}

func getRegionsBySize(sizeSlug string) {
	req, err := http.NewRequest("GET", "https://api.digitalocean.com/v2/sizes", nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("DO_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var sizes Sizes
	json.Unmarshal(body, &sizes)

	for _, size := range sizes.Sizes {
		if strings.ToLower(size.Slug) == strings.ToLower(sizeSlug) {
			fmt.Printf("Regions available for size '%s':\n", size.Slug)
			for _, region := range size.Regions {
				fmt.Println(region)
			}
			return
		}
	}

	fmt.Printf("Size '%s' not found.\n", sizeSlug)
}

// saveSSHKey saves the SSH key for the droplet in the specified folder
func saveSSHKey(dropletName string, privateKey, publicKey []byte) {
	keyDir := filepath.Join("keys", dropletName)
	err := os.MkdirAll(keyDir, 0700)
	if err != nil {
		log.Fatal(err)
	}

	privateKeyPath := filepath.Join(keyDir, "id_rsa")
	err = ioutil.WriteFile(privateKeyPath, privateKey, 0600)
	if err != nil {
		log.Fatal(err)
	}

	publicKeyPath := filepath.Join(keyDir, "id_rsa.pub")
	err = ioutil.WriteFile(publicKeyPath, publicKey, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("SSH keys saved for droplet '%s':\nPrivate key: %s\nPublic key: %s\n", dropletName, privateKeyPath, publicKeyPath)
}

func main() {
	opts := &options{}

	flagSet := goflags.NewFlagSet()
	flagSet.SetDescription("DigitalOcean Droplet Management")

	flagSet.BoolVar(&opts.listDroplets, "drops", false, "List all droplets")
	flagSet.BoolVar(&opts.deleteAll, "dry", false, "Dry run: delete all deployed droplets")
	flagSet.StringVar(&opts.fleet, "fleet", "droplet", "Name of the fleet (default: droplet)")
	flagSet.IntVar(&opts.amount, "amount", 2, "Specify the number of droplets to create, up to a maximum of 25.")
	flagSet.BoolVar(&opts.sizes, "sizes", false, "List all available sizes at DigitalOcean")
	flagSet.StringVar(&opts.size, "size", "", "Specify the size to check available regions")

	if err := flagSet.Parse(); err != nil {
		log.Fatalf("Could not parse flags: %s\n", err)
	}

	if opts.sizes {
		listSizes()
		return
	}

	if opts.listDroplets {
		listDroplets()
		return
	}

	if opts.deleteAll {
		deleteDroplets()
		return
	}

	if opts.size != "" {
		getRegionsBySize(opts.size)
		return
	}

	if opts.amount > 0 {
		if opts.amount > 25 {
			log.Fatalf("Cannot create more than 25 droplets.")
		}

		createDroplet(opts.fleet, opts.amount)
		return
	}

	// createDroplet("droplet", 1)
}
