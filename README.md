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
    - a droplet generator, remover, lister
    - export DO_API_KEY=your_api_key
    - help
        - go run main.go 			= create a new droplet
        - go run main.go -drops 	= list all droplets
        - go run main.go -dry 		= delete all deployed droplets
