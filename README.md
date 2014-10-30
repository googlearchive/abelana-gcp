# Cloud-Abelana-Go

![logo](https://cloud-abelana-go.github.com/images/image05.png)
![Architecture](https://cloud-abelana-go.github.com/images/image00.png)

A description of what this project does and who it serves.

Include authorship, support contact and release information.


## Project setup, installation, and configuration

How do I, as a developer, start working on the project?

1. What dependencies does it have (where are they expressed) and how do I install them?
1. Can I see the project working before I change anything?

1. How we setup Redis
  * Use one click install to get us 3 redis instances (master - 2 slaves)
  * For each instance, set it to restart automatically
  * get the Internal & External IP address for the Master
  * Add a backup cron job to backup the db every 15 minutes.
  * (Optional) Add firewall entries to all me to access from my development system
  * Add IP's to app config files

## Testing

How do I run the project's automated tests?

* Unit Tests

* Integration Tests


## Deploying

### How to setup the deployment environment

* Addons, packages, or other dependencies required for deployment.
* Required environment variables or credentials not included in git.
* Monitoring services and logging.

### How to deploy


## Troubleshooting & useful tools

### Examples of common tasks


```java
static SecureRandom sr = new SecureRandom();

byte[] android = new byte[32];
byte[] server = new byte[32];
byte[] password = new byte[32];

sr.nextBytes(android);
sr.nextBytes(server);

System.out.println("android:  "+ Base64.encodeToString(android, Base64.NO_PADDING | Base64.URL_SAFE));
System.out.println("server:   "+ Base64.encodeToString(server, Base64.NO_PADDING | Base64.URL_SAFE));

for(int i = 0; i<32; i++) password[i] = (byte) (android[i] ^ server[i]);
System.out.println("password: "+ Base64.encodeToString(key, Base64.NO_PADDING | Base64.URL_SAFE));
```

#### Changing the password on the p12 file:
* openssl pkcs12 -in < key.p12 > -nocerts -passin pass:notasecret -nodes -out /tmp/me.pem
* openssl pkcs12 -export -in /tmp/me.pem -out < mykey.p12 > -name privatekey -passout < New Passphrase >

e.g.
* How to make curl requests while authenticated via oauth.
* How to monitor background jobs.
* How to run the app through a proxy.

### Suggested Reading
* https://cloud.google.com/appengine/kb/general#static-ip

## Contributing changes

* See [CONTRIB.md](CONTRIB.md)


## Licensing

* See [LICENSE](LICENSE)
