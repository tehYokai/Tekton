// File: main.go
// export DO_API_KEY=your_api_key
// go run main.go 			= create a new droplet
// go run main.go --drops 	= list all droplets
// go run main.go --dry 		= delete all deployed droplets
// go run main.go --fleet 5 = creates 5 droplets
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
	"strconv"

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
}

type CreateDroplet struct {
	Name    string   `json:"name"`
	Region  string   `json:"region"`
	Size    string   `json:"size"`
	Image   string   `json:"image"`
	SSHKeys []string `json:"ssh_keys"`
}

type Droplets struct {
	Droplets []Droplet `json:"droplets"`
}

type CreateDropletResponse struct {
	Droplet Droplet `json:"droplet"`
}

type SSHKey struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

type options struct {
	listDroplets bool
	deleteAll    bool
	fleet        int
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

	for _, droplet := range droplets.Droplets {
		fmt.Printf("ID: %d, Name: %s, Region: %s, Image: %s\n", droplet.ID, droplet.Name, droplet.Region.Slug, droplet.Image.Slug)
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
		}
	}

	fmt.Println("All droplets deleted successfully.")
}

func createDroplet(name string) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

	// Generate and write private key as PEM
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privBLK := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	privPEM := pem.EncodeToMemory(&privBLK)
	if err = ioutil.WriteFile("id_rsa", privPEM, 0600); err != nil {
		log.Fatal(err)
	}

	// Generate and write public key
	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatal(err)
	}

	pubBytes := ssh.MarshalAuthorizedKey(pub)
	if err = ioutil.WriteFile("id_rsa.pub", pubBytes, 0644); err != nil {
		log.Fatal(err)
	}

	// Create new SSH key on DO account
	sshKey := SSHKey{
		Name:      name,
		PublicKey: string(pubBytes),
	}

	b, err := json.Marshal(sshKey)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", "https://api.digitalocean.com/v2/account/keys", bytes.NewBuffer(b))
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

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	keyID := strconv.Itoa(int(result["ssh_key"].(map[string]interface{})["id"].(float64)))

	// Create new Droplet with SSH key
	droplet := CreateDroplet{
		Name:    name,
		Region:  "nyc1",
		Size:    "s-1vcpu-1gb",
		Image:   "ubuntu-20-04-x64",
		SSHKeys: []string{keyID},
	}

	b, err = json.Marshal(droplet)
	if err != nil {
		log.Fatal(err)
	}

	req, err = http.NewRequest("POST", "https://api.digitalocean.com/v2/droplets", bytes.NewBuffer(b))
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

	body, _ = ioutil.ReadAll(resp.Body)

	var dropletResponse CreateDropletResponse
	json.Unmarshal(body, &dropletResponse)

	fmt.Printf("Created droplet with ID: %d\n", dropletResponse.Droplet.ID)
}

func main() {
	opts := &options{}

	flagSet := goflags.NewFlagSet()
	flagSet.SetDescription("DigitalOcean Droplet Management")

	flagSet.BoolVar(&opts.listDroplets, "drops", false, "List all droplets")
	flagSet.BoolVar(&opts.deleteAll, "dry", false, "Dry run: delete all deployed droplets")
	flagSet.IntVar(&opts.fleet, "fleet", 0, "Create a fleet of droplets. Specify the number of droplets to create, up to a maximum of 25.")

	if err := flagSet.Parse(); err != nil {
		log.Fatalf("Could not parse flags: %s\n", err)
	}

	if opts.listDroplets {
		listDroplets()
		return
	}

	if opts.deleteAll {
		deleteDroplets()
		return
	}

	if opts.fleet > 0 {
		if opts.fleet > 25 {
			log.Fatalf("Cannot create more than 25 droplets.")
		}

		for i := 0; i < opts.fleet; i++ {
			createDroplet(fmt.Sprintf("example-%d", i+1))
		}
		return
	}

	createDroplet("example")
}
