# Tekton
 The Ancient Greek noun tektōn is a common term for an artisan/craftsman, in particular a carpenter, woodworker, or builder. The term is frequently contrasted with an ironworker, or smith and stone-worker or mason.

## Tools
- vpn-rotator
    - sh version works
        - V1 works, switches VPN connection every 60 seconds
            - download your .ovpn's and put them in ./ovpn directory
        - V2 todo
            - V2 will not allow vpn-less connections while switching VPN endpoints
    - deno version does not work

- drop-gen
    - go run main.go -h                
```bash
DigitalOcean Droplet Management

Flags:
   -drops         List all droplets
   -dry           Dry run: delete all deployed droplets
   -fleet string  Name of the fleet (default: droplet) (default "droplet")
   -amount int    Specify the number of droplets to create, up to a maximum of 25. (default 2)
   -sizes         List all available sizes at DigitalOcean
   -size string   Specify the size to check available regions
```
    - todo's:
        - createDroplet()
            - check for drops.json first
                - based on that increment [name] if the [name] is already in drops.json
        - install/setup tools/env on droplet
    - ideas:
        - interactive terminal for general setup


