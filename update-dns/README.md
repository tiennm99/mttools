# update-dns

Update CloudFlare DNS records with dynamic IP when your IP is behind router's IP.

## How to use this script

```bash
git clone https://github.com/tiennm99/update-dns
cd update-dns
nano main.sh
# Edit value in the script
./main.sh
```

You can also use the script with cronjob to update your IP every 5 minutes

```bash
crontab -e
# Add the following line to the file
*/5 * * * * /path/to/main.sh
```

## __Note:__
You may need to config your router to forward the ports you wanted to the exact machine.
If you have problems, let's discuss in the discussions section.
