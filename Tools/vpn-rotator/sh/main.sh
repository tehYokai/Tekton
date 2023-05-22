#!/bin/bash

# curl -s https://ifconfig.co

# List of OpenVPN configuration files
# https://account.protonvpn.com/downloads#openvpn-configuration-files
config_dir="./ovpn"
servers=("")

# username / password
# USERNAME=""
# PASSWORD=""
# Source the .env file
source .env

# Function to handle Ctrl+C
function handle_signal() {
    echo -e "\nReceived Ctrl+C. Disconnecting from all VPN connections..."
    pkill openvpn
    exit 0
}

# IP address before connecting to VPN
echo "Your IP address before connecting to VPN is: $(curl -s https://ifconfig.co)"

# Register the Ctrl+C signal handler
trap handle_signal SIGINT

while true; do # can be made easier by merging config_dir and servers, a use of a combination of a loop and the find command
    for server in "${servers[@]}"; do
        config_file="${config_dir}/${server}.ovpn"
        echo "Connecting to ${config_file}"

        # Run OpenVPN with credentials provided
        openvpn --config "$config_file" --daemon --auth-user-pass <(printf "%s\n%s\n" "$USERNAME" "$PASSWORD")
        
        # Wait for the VPN connection to be established
        sleep 2

        # Retrieve the VPN IP address
        # vpn_ip=$(ifconfig tun0 | awk '/inet / {print $2}')
        # echo "Your VPN IP is: $vpn_ip"
        echo "Your VPN IP is: $(curl -s https://ifconfig.co)"

        # Wait for 60 seconds before disconnecting
        sleep 60

        echo "Disconnecting from ${config_file}"
        pkill openvpn
        sleep 2
    done
done
