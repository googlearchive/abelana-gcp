## TODO List - Missing Code
* Endpoints
  * [x] GetMyProfile
  * [ ] Follow (given email)
  * [ ] AddFollower (partially implemented)
  * [x] FProfile - Get a specific followers entries only
  * [x] GetFollowers (just needs to be enabled)
  * [ ] GetFollowers add Names.
  * [x] GetFollower (just needs to be enabled)
  * [x] Flag
  * [ ] Wipeout
  * [ ] Register GCM
  * [ ] Unregister GCM
  * [ ] Import
    * [ ] Facebook
    * [ ] G+
    * [ ] Yahoo

* [ ] Redis Connection Timeouts (if any)
* [ ] FIXME Backdoor is currently enabled
* [ ] Redis - Batch LPUSH's
* [ ] FIXME tokens.Login should use GitKit's VerifyToken
* [ ] FIXME Remove HalfPW from AccToken
* [ ] FIXME tokens.VerifyToken isn't using Certs.
* [ ] tokens.Login should verify Audience is correct
* [ ] Backup the Redis DB to it's own bucket on CloudStorage every 15 minutes.
* [ ] Move configuration from static info to a config file
