# Tekton
 The Ancient Greek noun tektōn is a common term for an artisan/craftsman, in particular a carpenter, woodworker, or builder. The term is frequently contrasted with an ironworker, or smith and stone-worker or mason.

## Tools
- manager/manager.go
    - `go run tg2.go -time 2m -task "coding manager"`
    - it will create a tasks.csv file and append on what date for how long have you been doing the task for
    - at the end it will speak out words - like an alarm
- vpn-rotator
    - sh version works
        - V1 works, switches VPN connection every 60 seconds
            - download your .ovpn's and put them in ./ovpn directory
        - V2 todo
            - V2 will not allow vpn-less connections while switching VPN endpoints
    - deno version does not work

- drop-gen has evolved beyond the scope of tekton repository, it's a pretty solid codebase of 3000+ lines of code now, it's basic usage has been completed and scaled to a certain use, but i'm yet to add so much more. Drop-gen is but one of the 'idea' versions, you can use it as an inspiration.


