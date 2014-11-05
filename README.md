# Cloud-Abelana-Go
<!--
![Abelana](https://abelana-gcp.github.com/images/0001.0001.png)
![logo](https://abelana-gcp.github.com/images/image05.png)
-->
![Architecture](https://GoogleCloudPlatform.github.io/abelana-gcp/images/architecture.png)

Abelana (means "Share" in Zulu) is a project that allows users to take photos and share them with
their friends.

This project contains sample code for the [Google Cloud Platform](https://cloud.google.com/).
A companion [Android Client](https://github.com/GoogleCloudPlatform/Abelana-Android) is helpful to see.

The Cloud apps were written by Les Vogel and Francesc Campoy Flores. For questions and comments,
please join the [Google Group](https://groups.google.com/forum/#!forum/abelana-app).

Disclaimer: This sample application is for educational purposes only and is not a Google product or service.

# This is a work in progress
### It is incomplete
### Please be patient!
The goal of this project is to help you learn how to create your applications in the cloud. We will
be supporting this project with videos and additional material.

## Project setup, installation, and configuration
### (Not yet enough to get it running)

How do I, as a developer, bring this project up as my own Google Cloud Platform project?

1. Grab the project from GitHub
  * Following the Go guidelines, the project belongs in your GOPATH (ours is ~/go).
  * The files go into $GOPATH/src/github.com/GoogleCloudPlatform/abelana-gcp.
  * There are 3 Appe Engine applications:
    * **default** - the main website, supports identity-toolkit sign-on. (We took their sample).
    * **endpoints** - Where most of the action occurs.
    * **notice** - Takes [Object Change Notifications](https://cloud.google.com/storage/docs/object-change-notification) from Google Cloud Storage and connects with the Image magick for resizing and re-encoding. Then uploads to GCS, and notifies endpoints.
  * There is also a third party directory of code we modified.
  * The GCE app
    * **imagemagick** - the docker component for hosting imagemagick.
  * [Redis](http://redis.io/) -- Not much there, you should modify the config files for your instance yourself.

1. Create a Cloud Project
  * Using the [Developer Console](https://console.developers.google.com/project):
    * Set up billing.
    * Create the project.

1. Click on **Permissions**.
  * Take note of the App Engine Service Account. You'll need this later.

1. Click on **Credentials**. You'll need the following:
  * Client ID for Android application
    * Please see [this page](https://developers.google.com/+/mobile/android/getting-started#step_1_enable_the_google_api), you
    only need step 1 #5.
  * Client ID for web application.
    * redirect URI's should include:
        * https://localhost/callback
        * https//<your-appengine-project>.appspot.com/gitkit

  * Service Account
    * Generate and download a new P12 key.

  * Public API Access
    * Key for browser applications.
    * Key for server applications.

1. Click on **APIs**. You will need the following:
  * Google Cloud Storage
  * Google Cloud Storage JSON API
  * Google Compute Engine
  * Google Compute Engine Autoscaler API
  * Google Compute Engine Instance Group Manager API
  * Google Compute Engine Instance Group Updater API
  * Google Compute Engine Instance Groups API
  * Google+ API
  * Identity Toolkit API

1. Details for the [Android Client](https://github.com/GoogleCloudPlatform/Abelana-Android)
  * How we create create the secretKey that resides on Android, used to access Cloud Storage:
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
    System.out.println("passphrase: "+ Base64.encodeToString(key, Base64.NO_PADDING | Base64.URL_SAFE));
    ```

    * Changing the password on the p12 file:
        * `openssl pkcs12 -in < key.p12 > -nocerts -passin pass:notasecret -nodes -out /tmp/me.pem`
        * `openssl pkcs12 -export -in /tmp/me.pem -out < mykey.p12 > -name privatekey -passout < New Passphrase > `

e.g.
* How to make curl requests while authenticated via oauth.
* How to monitor background jobs.
* How to run the app through a proxy.

1. [Identity Toolkit]()
  * TBD

1. Google Cloud Storage
  * TBD

1. App Engine Modules
   * In **abelana-gcp/endpoints**, create a folder called **private**
   * Create **private/gitkit-server-config.json**
   ```json
   {
  "clientId" : "41652380zzzz-xxxxxxxxxxxxx.apps.googleusercontent.com",
  "serverApiKey" : "yyyyyyyyyyyyyyyyyyyyyyyy",
  "widgetUrl" : "https//<your-appengine-project>.appspot.com/gitkit",
  "cookieName" : "gtoken"
}
```
   * Create **private/abelana-config.json**
   ```json
   {
    "AuthEmail" : "<Service Account Email>@developer.gserviceaccount.com",
    "ProjectID" : "<YOUR PROJECT ID>",
    "Bucket" : "<Your Upload bucket>",
    "RedisPW" : "<YOUR REDIS PASSWORD>",
    "Redis" : "<IP OF YOUR REDIS INSTANCE>:6379",
    "TimelineBatchSize" : 100,
    "UploadRetries" : 5,
    "EnableBackdoor" : false,
    "EnableStubs" : false
}
   ```

1. [Redis](http://redis.io/):
  * Use one click install, to start, you only need 1 instance.
  * Use **gcloud instance ssh ...** to connect with your instance.
  * Edit /etc/redis/redis.config.
  * Add `requirepass "<YOUR REDIS PASSWORD>"`.

1. Image Magick

1. What dependencies does it have (where are they expressed) and how do I install them?

1. Can I see the project working before I change anything?

1. How we set up Redis:
  * Use one click install to get us 3 redis instances (master - 2 slaves).
  * For each instance, set it to restart automatically.
  * get the Internal & External IP address for the Master.
  * Add a backup cron job to backup the db every 15 minutes.
  * (Optional) Add firewall entries to all me to access from my development system.
  * Add IP's to app config files.

## Testing

How do I run the project's automated tests?

* Unit Tests

* Integration Tests


## Deploying

### How to set up the deployment environment

* Add-ons, packages, or other dependencies required for deployment.
* Required environment variables or credentials not included in git.
* Monitoring services and logging.

### How to deploy


## Troubleshooting & useful tools

### Examples of common tasks

e.g.
* How to make curl requests while authenticated via oauth.
* How to monitor background jobs.
* How to run the app through a proxy.

### Suggested Reading
* https://cloud.google.com/appengine/kb/general#static-ip
* https://cloud.google.com/storage/docs/authentication

## Contributing changes

* See [CONTRIB.md](CONTRIB.md)


## Licensing

* See [LICENSE](LICENSE)
