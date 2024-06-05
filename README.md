# pinger

a simple infrastructure tool to monitor other applications.

- Simple: can be configured by writing one config file
- Lightweight: should take no more than 20MBs of ram and run on very cheap hardware
- Powerful: It should be a primitive that is helpful in constructing other systems.

# Running the tool

1. Clone the repo

```bash
$ git clone git@github.com:aramtech/pinger.git
```

2. Compile the source

```bash
$ go build ./...
```

3. Generate a config file

```bash
$ ./pinger -config
```

4. Modify the config file to suit your needs

```json
{
  // A list of the apps to monitor
  "apps": [
    {
      "appName": "Delivery platform", // The application name
      "statusUrl": "http://fma.aramtech.ly/server/api/status", // Status url
      "checkInterval": "1m", // Check interval in ms
      "httpReporters": [
        // Alert handlers
        {
          "url": "https://easysms.devs.ly/sms/api", // Endpoint to hit if a service goes down
          "method": "POST",
          // The body sent with the http request
          "body": {
            "action": "send-sms",
            "api_key": "API_KEY",
            "unicode": 1,
            "to": "911974326",
            "sms": "delivery platform server is down!"
          }
        }
      ]
    }
  ]
}
```

Basically for each defined app, pinger will check the status of the application if the application responds with a none 200 response code
pinger will send http requests to the specified alert urls, the above config will monitor the service at the URL `https://fma.aramtech.ly/sms/api` and will send alerts if it the service goes down.
