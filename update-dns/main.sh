#!/bin/bash

# Configuration
CF_EMAIL="your@email.com"
CF_API_KEY="your_cloudflare_api_key"
ZONE_ID="your_zone_id"
RECORDS=("subdomain1.example.com" "subdomain2.example.com" "subdomain3.example.com")

# Get current IP address using upnpc
CURRENT_IP=$(upnpc -s | awk -F'= ' '/ExternalIPAddress/ { print $2 }')

# Function to update DNS record
update_record() {
    local record="$1"
    local record_id=$(curl -sX GET "https://api.cloudflare.com/client/v4/zones/$ZONE_ID/dns_records?type=A&name=$record" \
        -H "X-Auth-Email: $CF_EMAIL" \
        -H "Authorization: Bearer $CF_API_KEY" \
        -H "Content-Type: application/json" | jq -r '.result[0].id')

    if [ -n "$record_id" ]; then
        curl -sX PUT "https://api.cloudflare.com/client/v4/zones/$ZONE_ID/dns_records/$record_id" \
            -H "X-Auth-Email: $CF_EMAIL" \
            -H "Authorization: Bearer $CF_API_KEY" \
            -H "Content-Type: application/json" \
            --data "{\"type\":\"A\",\"name\":\"$record\",\"content\":\"$CURRENT_IP\"}"
        echo "Updated $record to $CURRENT_IP"
    else
        echo "Error: $record not found"
    fi
}

# Update DNS records
for record in "${RECORDS[@]}"; do
    update_record "$record"
done
