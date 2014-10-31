## TODO List - Missing Code
* Endpoints
  * [ ] Follow (given email)
  * [ ] AddFollower (partially implemented)
  * [ ] Wipeout
  * [ ] Register GCM
  * [ ] Unregister GCM
  * [ ] Import
    * [ ] Facebook
    * [ ] G+
    * [ ] Yahoo

* [ ] findFollows needs to be implemented
* [ ] Redis Connection Timeouts (if any)
* [ ] FIXME Backdoor is currently enabled
* [ ] FIXME tokens.Login should use GitKit's VerifyToken
* [ ] FIXME Remove HalfPW from AccToken
* [ ] tokens.Login should verify Audience is correct
* [ ] Backup the Redis DB to it's own bucket on CloudStorage every 15 minutes.
* [ ] Move configuration from static info to a config file
* [ ] Redis Password
* [ ] use RunInTransaction more...

## Later
* [ ] Redis - Batch LPUSH's
* [ ] Follow - Lookup email in GitKit instead of searching Datastore
* [ ] Switch to projection Queries
