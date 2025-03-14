# Change Log
All notable changes to this project are documented below.

The format is based on [keep a changelog](http://keepachangelog.com) and this project uses [semantic versioning](http://semver.org).

## [Unreleased]
### Added
- New Lua runtime functions to generate JWT tokens.
- New Lua runtime functions to hash data using RSA SHA256.
- Print max number of OS threads setting in server startup logs.

### Changed
- Log more information when authoritative match handlers receive too many data messages.
- Ensure storage writes and deletes are performed in a consistent order within each batch.
- Ensure wallet updates are performed in a consistent order within each batch.

### Fixed
- Storage write batches now correctly abort when any query in the batch fails.
- Rank cache correctly calculates record expiry times.
- Return correct response to group join operations when the user is already a member of the group.
- Fix query when selecting a page of leaderboard records around a user.

## [2.4.2] - 2019-03-25
### Added
- New programmatic console API for administrative server operations.
- Initial events subsystem with session start+end handlers.

### Changed
- Update GRPC (1.19.0), GRPC-Gateway (1.8.4), and Protobuf (1.3.0) dependencies.
- Use Go 1.12.1 as base Docker container image and native builds.

## [2.4.1] - 2019-03-08
### Added
- Strict validation of socket timeout configuration parameters.
- New Go runtime constants representing storage permissions.
- New runtime function to programmatically delete user accounts.
- Allow multiple config files to be read at startup and merged into a final server configuration.
- Storage listing operations can now disambiguate between listing system-owned data and listing all data.

### Changed
- Default maximum database connection lifetime is now 1 hour.
- Improved parsing of client IP and port for incoming requests and socket connections.
- WebSocket sessions no longer log the client IP and port number in error messages.
- Go and Lua server runtime startup log messages are now consistent.
- All schema and query statements that use the '1970-01-01 00:00:00' constant now specify UTC timezone.
- Storage write error message are more descriptive for when values must be encoded JSON objects.
- Storage listing operations now treat empty owner IDs as listing across all data rather than system-owned data.
- Storage write operations now return more specific error messages.

### Fixed
- CRON expressions for leaderboard and tournament resets now allow concurrent usage safely.
- Set console API gateway timeout to match connection idle timeout value.

## [2.4.0] - 2019-02-03
### Added
- New logging format option for Stackdriver Logging.
- New runtime function to immediately disconnect active sockets.
- New runtime function to kick arbitrary presences from streams.

### Fixed
- Fix return arguments for group user list results in Lua runtime function.
- Leaderboard records returned with a previous page cursor no longer errors.

## [2.3.2] - 2019-01-17
### Fixed
- Set gateway timeout to match idle timeout value.
- Reliably release database resources before moving from one query to the next.
- Unlock GPGS certs cache in social client.

## [2.3.1] - 2019-01-04
### Added
- Make authoritative match join attempt marker deadline configurable.

### Changed
- Improve db transaction semantics with batch wallet updates.

### Fixed
- Initialize registration of deferred messages sent from authoritative matches.
- Early cancel Lua authoritative match context when match initialization fails.
- Update decoding of Steam authentication responses to correctly unwrap payload. Thanks @nielslanting
- Parse Steam Web API response errors when authenticating Steam tokens.

## [2.3.0] - 2018-12-31
### Added
- WebSocket connections can now send Protobuf binary messages.
- Lua runtime tournament listings now return duration, end active, and end time fields.
- Lua runtime tournament end hooks now contain duration, end active, and end time fields.
- Lua runtime tournament reset hooks now contain duration, end active, and end time fields.
- New configuration flag for maximum number of concurrent join requests to authoritative matches.
- New runtime function to kick users from a group.
- Clients that send data to an invalid match ID will now receive an uncollated error.
- The logger now supports optional log file rotation.
- Go runtime authoritative matches now also print Match IDs in log lines generated within the match.
- Email authentication client requests can authenticate with username/password instead of email/password.
- Email authentication server runtime calls can authenticate with username/password instead of email/password.
- New authoritative match dispatcher function to defer message broadcasts until the end of the tick.
- New runtime function to retrieve multiple user accounts by user ID.
- Send notifications to admins of non-open groups when a user requests to join.
- Send notifications to users when their request to join a group is accepted.
- New configuration flag for presence event buffer size.

### Changed
- Replace standard logger supplied to the Go runtime with a more powerful interface.
- Rename stream 'descriptor' field to 'subcontext' to avoid protocol naming conflict.
- Rename Facebook authentication and link 'import' field to avoid language keyword conflict.
- Rejoining a match the user is already part of will now return the match label.
- Allow tournament joins before the start of the tournament active period.
- Authoritative matches now complete their stop phase faster to avoid unnecessary processing.
- Authoritative match join attempts now have their own bounded queue and no longer count towards the match call queue limit.
- Lua runtime group create function now sets the correct default max size if one is not specified.
- Improve socket session close semantics.
- Session logging now prints correct remote address if available when the connection is through a proxy.
- Authoritative match join attempts now wait until the handler acknowledges the join before returning to clients.

### Fixed
- Report correct execution mode in Lua runtime after hooks.
- Use correct parameter type for creator ID in group update queries.
- Use correct parameter name for lang tag in group update queries.
- Do not allow users to send friend requests to the root user.
- Tournament listings now report correct active periods if the start time is in the future.
- Leaderboard and tournament reset runtime callbacks now receive the correct reset time.
- Tournament end runtime callbacks now receive the correct end time.
- Leaderboard and tournament runtime callbacks no longer trigger twice when time delays are observed.
- Check group max allowed user when promoting a user.
- Correct Lua runtime decoding of stream identifying parameters.
- Correctly use optional parameters when they are passed to group creation operations.
- Lua runtime operations now observe context cancellation while waiting for an available Lua instance.
- Correctly list tournament records when the tournament has no end time defined.

## [2.2.1] - 2018-11-20
### Added
- New duration field in the tournament API.

### Fixed
- Set friend state correctly when initially adding friends.
- Allow tournaments to be created to start in the future but with no end time.
- Join events on tournaments with an end time set but no reset now allow users to submit scores.

## [2.2.0] - 2018-11-11
### Added
- New runtime function to send raw realtime envelope data through streams.

### Changed
- Improve error message on database errors raised during authentication operations.
- Set new default of 100 maximum number of open database connections.
- Friendship state is no longer offset by one when sent to clients.
- Group membership state is no longer offset by one when sent to clients.
- Set new default metrics report frequency to 60 seconds.

### Fixed
- Account update optional inputs are not updated unless set in runtime functions.
- Fix boolean logic with context cancellation in single-statement database operations.

## [2.1.3] - 2018-11-02
### Added
- Add option to skip virtual wallet ledger writes if not needed.

### Changed
- Improved error handling in Lua runtime custom SQL function calls.
- Authoritative match join attempts are now cancelled faster when the client session closes.

### Fixed
- Correctly support arbitrary database schema names that may contain special characters.

## [2.1.2] - 2018-10-25
### Added
- Ensure runtime environment values are exposed through the Go runtime InitModule context.

### Changed
- Log more error information when InitModule hooks from Go runtime plugins return errors.
- Preserve order expected in match listings generated with boosted query terms.

### Fixed
- Improve leaderboard rank re-calculation when removing a leaderboard record.

## [2.1.1] - 2018-10-21
### Added
- More flexible query-based filter when listing realtime multiplayer matches.
- Runtime function to batch get groups by group ID.
- Allow authoritative match join attempts to carry metadata from the client.

### Changed
- Improved cancellation of ongoing work when clients disconnect.
- Improved validation of dispatcher broadcast message filters.
- Set maximum size of authoritative match labels to 2048 bytes.

### Fixed
- Use leaderboard expires rather than end active IDs with leaderboard resets.
- Better validation of tournament duration when a reset schedule is set.
- Set default matchmaker input query if none supplied with the request.
- Removed a possible race condition when session ping backoff triggers concurrently with a timed ping.
- Errors returned by InitModule hooks from Go runtime plugins will now correctly halt startup.

## [2.1.0] - 2018-10-08
### Added
- New Go code runtime for custom functions and authoritative match handlers.
- New Tournaments feature.
- Runtime custom function triggers for leaderboard and tournament resets.
- Add Lua runtime AES-256 util functions.
- Lua runtime token generator function now returns a second value representing the token's expiry.
- Add local cache for in-memory storage to the Lua runtime.
- Graceful server shutdown and match termination.
- Expose incoming request data in runtime after hooks.

### Changed
- Improved Postgres compatibility on TIMESTAMPTZ types.

### Fixed
- Correctly merge new friend records when importing from Facebook.
- Log registered hook names correctly at startup.

## [2.0.3] - 2018-08-10
### Added
- New "bit32" backported module available in the code runtime.
- New code runtime function to create MD5 hashes.
- New code runtime function to create SHA256 hashes.
- Runtime stream user list function now allows filtering hidden presences.
- Allow optional request body compression on all API requests.

### Changed
- Reduce the frequency of socket checks on known active connections.
- Deleting a record from a leaderboard that does not exist now succeeds.
- Notification listings use a more accurate timestamp in cacheable cursors.
- Use "root" as the default database user if not specified.

### Fixed
- Runtime module loading now correctly handles paths on non-UNIX environments.
- Correctly handle blocked user list when importing friends from Facebook.

## [2.0.2] - 2018-07-09
### Added
- New configuration option to adjust authoritative match data input queue size.
- New configuration option to adjust authoritative match call queue size.
- New configuration options to allow listening on IPv4/6 and a particular network interface.
- Authoritative match modules now support a `match_join` callback that triggers when users have completed their join process.
- New stream API function to upsert a user presence.
- Extended validation of Google signin tokens to handle different token payloads.
- Authoritative match labels can now be updated using the dispatcher's `match_label_update` function.

### Changed
- Presence list in match join responses no longer contains the user's own presence.
- Presence list in channel join responses no longer contains the user's own presence.
- Socket read/write buffer sizes are now set based on the `socket.max_message_size_bytes` value.
- Console GRPC port now set relative to `console.port` config value.

## [2.0.1] - 2018-06-15
### Added
- New timeout option to HTTP request function in the code runtime.
- Set QoS settings on client outgoing message queue.
- New runtime pool min/max size options.
- New user ban and unban functions.
- RPC functions triggered by HTTP GET requests now include any custom query parameters.
- Authoritative match messages now carry a receive timestamp field.
- Track new metrics for function calls, before/after hooks, and internal components.

### Changed
- The avatar URL fields in various domain objects now support up to 512 characters for FBIG.
- Runtime modules are now loaded in a deterministic order.

### Fixed
- Add "ON DELETE CASCADE" to foreign key user constraint on wallet ledger.

## [2.0.0] - 2018-05-14

This release brings a large number of changes and new features to the server. It cannot be upgraded from v1.0 - reach out for help to upgrade.

### Added
- Authenticate functions can now be called from the code runtime.
- Use opencensus for server metrics. Add drivers for Prometheus and Google Cloud Stackdriver.
- New API for users to subscribe to status update events from other users online.
- New API for user wallets to store and manage virtual currencies.
- Realtime multiplayer supports authoritative matches with a handler and game loop on the server.
- Matches can be listed on the server for "room-based" matchmaker logic.
- "run_once" function to execute logic at startup with the code runtime.
- Variables can be passed into the server for environment configuration.
- Low level streams API for advanced distributed use cases.
- New API for export and delete of users for GDPR compliance.

### Changed
- Split the server protocol into request/response with GRPC or HTTP1.1+JSON (REST) and WebSockets or rUDP.
- The command line flags of the server have changed to be clearer and more explicit.
- Authenticate functions can now take username as an input at account create time.
- Use TIMESTAMPTZ for datetimes in the database.
- Use JSONB for objects stored in the database.
- Before/after hooks changed to distinguish between req/resp and socket messages.
- Startup messages are more concise.
- Log messages have been updated to be more useful in development.
- Stdlib for the code runtime uses "snake_case" consistently across variables and function names.
- The base image for our Docker images now uses Alpine Linux.

### Fixed
- Build dependencies are now vendored and build system is simplified.
- Database requests for transaction retries are handled automatically.

### Removed
- The storage engine no longer needs a "bucket" field as a namespace. It was redundant.
- Leaderboard haystack queries did not perform well and need a redesign.
- IAP validation removed until it can be integrated with the virtual wallet system.

---

## [1.4.1] - 2018-03-30
### Added
- Allow the server to handle SSL termination of client connections although NOT recommended in production.
- Add code runtime hook for IAP validation messages.

### Changed
- Update social sign-in code for changes to Google's API.
- Migrate code is now cockroach2 compatible.

### Fixed
- Fix bitshift code in rUDP protocol parser.
- Fix incorrect In-app purchase setup availability checks.
- Cast ID in friend add queries which send notifications.
- Expiry field in notifications now stored in database write.
- Return success if user is re-added who is already a friend.

## [1.4.0] - 2017-12-16
### Changed
- Nakama will now log an error and refuse to start if the schema is outdated.
- Drop unused leaderboard 'next' and 'previous' fields.
- A user's 'last online at' field now contains a current UTC milliseconds timestamp if they are currently online.
- Fields that expect JSON content now allow up to 32kb of data.

### Fixed
- Storage remove operations now ignore records that don't exist.

## [1.3.0] - 2017-11-21
### Added
- Improve graceful shutdown behaviour by ensuring the server stops accepting connections before halting other components.
- Add User-Agent to the default list of accepted CORS request headers.
- Improve how the dashboard component is stopped when server shuts down.
- Improve dashboard CORS support by extending the list of allowed request headers.
- Server startup output now contains database version string.
- Migrate command output now contains database version string.
- Doctor command output now contains database version string.

### Changed
- Internal operations exposed to the script runtime through function bindings now silently ignore unknown parameters.

### Fixed
- Blocking users now works correctly when there was no prior friend relationship in place.
- Correctly assign cursor data in paginated leaderboard records list queries.
- Improve performance of user device login operations.

## [1.2.0] - 2017-11-06
### Added
- New experimental rUDP socket protocol option for client connections.
- Accept JSON payloads over WebSocket connections.

### Changed
- Use string identifiers instead of byte arrays for compatibility across Lua, JSON, and client representations.
- Improve runtime hook lookup behaviour.

### [1.1.0] - 2017-10-17
### Added
- Advanced Matchmaking with custom filters and user properties.

### Changed
- Script runtime RPC and HTTP hook errors now return more detail when verbose logging is enabled.
- Script runtime invocations now use separate underlying states to improve concurrency.

### Fixed
- Build system no longer passes flags to Go vet command.
- Haystack leaderboard record listings now return correct results around both sides of the pivot record.
- Haystack leaderboard record listings now return a complete page even when the pivot record is at the end of the leaderboard.
- CRON expression runtime function now correctly uses UTC as the timezone for input timestamps.
- Ensure all runtime 'os' module time functions default to UTC timezone.

## [1.0.2] - 2017-09-29
### Added
- New code runtime function to list leaderboard records for a given set of users.
- New code runtime function to list leaderboard records around a given user.
- New code runtime function to execute raw SQL queries.
- New code runtime function to run CRON expressions.

### Changed
- Handle update now returns a bad input error code if handle is too long.
- Improved handling of accept request headers in HTTP runtime script invocations.
- Improved handling of content type request headers in HTTP runtime script invocations.
- Increase default maximum length of user handle from 20 to 128 characters.
- Increase default maximum length of device and custom IDs from 64 to 128 characters.
- Increase default maximum length of various name, location, timezone, and other free text fields to 255 characters.
- Increase default maximum length of storage bucket, collection, and record from 70 to 128 characters.
- Increase default maximum length of topic room names from 64 to 128 characters.
- Better error responses when runtime function RPC or HTTP hooks fail or return errors.
- Log a more informative error message when social providers are unreachable or return errors.

### Fixed
- Realtime notification routing now correctly resolves connected users.
- The server will now correctly log a reason when clients disconnect unexpectedly.
- Use correct wire format when sending live notifications to clients.

## [1.0.1] - 2017-08-05
### Added
- New code runtime functions to convert UUIDs between byte and string representations.

### Changed
- Improve index selection in storage list operations.
- Payloads in `register_before` hooks now use `PascalCase` field names and expose correctly formatted IDs.
- Metadata regions in users, groups, and leaderboard records are now exposed to the code runtime as Lua tables.

### Fixed
- The code runtime batch user update operations now process correctly.

## [1.0.0] - 2017-08-01
### Added
- New storage partial update feature.
- Log warn messages at startup when using insecure default parameter values.
- Add code runtime function to update groups.
- Add code runtime function to list groups a user is part of.
- Add code runtime function to list users who're members of a group.
- Add code runtime function to submit a score to a leaderboard.
- Send in-app notification on friend request.
- Send in-app notification on friend request accept.
- Send in-app notification when a Facebook friend signs into the game for the first time.
- Send in-app notification to group admins when a user requests to join a private group.
- Send in-app notification to the user when they are added to a group or their request to join a private group is accepted.
- Send in-app notification to the user when someone wants to DM chat.

### Changed
- Use a Lua table with content field when creating new notifications.
- Use a Lua table with metadata field when creating new groups.
- Use a Lua table with metadata field when updating a user.
- Updated configuration variable names. The most important one is `DB` which is now `database.address`.
- Moved all `nakamax` functions into `nakama` runtime module.
- An invalid config file or invalid cmdflag now prevents the server from startup.
- A matchmake token now expires after 30 instead of 15 seconds.
- The code runtime `os.date()` function now returns correct day of year.
- The code runtime context passed to function hooks now use PascalCase case in fields names. For example `context.user_id` is now `context.UserId`.
- Remove `admin` sub-command.
- A group leave operation now returns a specific error code when the last admin attempts to leave.
- A group self list operations now return the user's membership state with each group.

## [1.0.0-rc.1] - 2017-07-18
### Added
- New storage list feature.
- Ban users and create groups from within the code runtime.
- Update users from within the code runtime.
- New In-App Purchase validation feature.
- New In-App Notification feature.

### Changed
- Run Facebook friends import after registration completes.
- Adjust command line flags to be follow pattern in the config file.
- Extend the server protocol to be batch-orientated for more message types.
- Update code runtime modules to use plural function names for batch operations.
- The code runtime JSON encoder/decoder now support root level JSON array literals.
- The code runtime storage functions now expect and return Lua tables for values.
- Login attempts with an ID that does not exist will return a new dedicated error code.
- Register attempts with an ID that already exists will return a new dedicated error code.

### Fixed
- The runtime code for the after hook message was set to "before" incorrectly.
- The user ID was not passed into the function context in "after" authentication messages.
- Authentication messages required hook names which began with "." and "\_".
- A device ID used in a link message which was already in use now returns "link in use" error code.

## [0.13.1] - 2017-06-08
### Added
- Runtime Base64 and Base16 conversion util functions.

### Fixed
- Update storage write permissions validation.
- Runtime module path must derive from `--data-dir` flag value.
- Fix parameter mapping in leaderboard haystack query.

## [0.13.0] - 2017-05-29
### Added
- Lua script runtime for custom code.
- Node status now also reports a startup timestamp.
- New matchmaking feature.
- Optionally send match data to a subset of match participants.
- Fetch users by handle.
- Add friend by handle.
- Filter by IDs in leaderboard list message.
- User storage messages can now set records with public read permission.

### Changed
- The build system now suffixes Windows binaries with `exe` extension.

### Fixed
- Set correct initial group member count when group is created.
- Do not update group count when join requests are rejected.
- Use cast with leaderboard BEST score submissions due to new strictness in database type conversion.
- Storage records can now correctly be marked with no owner (global).

## [0.12.2] - 2017-04-22
### Added
- Add `--logtostdout` flag to redirect log output to console.
- Add build rule to create Docker release images.

### Changed
- Update Zap logging library to latest stable version.
- The `--verbose` flag no longer alters the logging output to print to both terminal and file.
- The log output is now in JSON format.
- Update the healthcheck endpoint to be "/" (root path) of the main server port.

### Fixed
- Fix a race when the heartbeat ticker might not be stopped after a connection is closed.

## [0.12.1] - 2017-03-28
### Added
- Optionally allow JSON encoding in user login/register operations and responses.

### Changed
- Improve user email storage and comparison.
- Allow group batch fetch by both ID and name.
- Increase heartbeat server time precision.
- Rework the embedded dashboard.
- Support 64 characters with `SystemInfo.deviceUniqueIdentifier` on Windows with device ID link messages.

### Fixed
- Fix Facebook unlink operation.

## [0.12.0] - 2017-03-19
### Added
- Dynamic leaderboards feature.
- Presence updates now report the user's handle.
- Add error codes to the server protocol.

### Changed
- The build system now strips up to current dir in recorded source file paths at compile.
- Group names must now be unique.

### Fixed
- Fix regression loading config file.

## [0.11.3] - 2017-02-25
### Added
- Add CORS headers for browser games.

### Changed
- Update response types to realtime match create/join operations.

### Fixed
- Make sure dependent build rules are run with `relupload` rule.
- Fix match presence list generated when joining matches.

## [0.11.2] - 2017-02-17
### Added
- Include Dockerfile and Docker instructions.
- Use a default limit in topic message listings if one is not provided.
- Improve log messages in topic presence diff checks.
- Report self presence in realtime match create and join.

### Changed
- Improve warn message when database is created in migrate subcommand.
- Print database connections to logs on server start.
- Use byte slices with most database operations.
- Standardize match presence field names across chat and realtime protocol.
- Improve concurrency for closed sockets.

### Fixed
- Enforce concurrency control on outgoing socket messages.
- Fix session lookup in realtime message router.
- Fix input validation when chat messages are sent.
- Fix how IDs are handled in various login options.
- Fix presence service shutdown sequence.
- More graceful handling of session operations while connection is closed.
- Fix batch user fetch query construction.
- Fix duplicate leaves reported in topic presence diff messages.

## [0.11.1] - 2017-02-12
### Changed
- Server configuration in dashboard is now displayed as YAML.
- Update server protocol to simplify presence messages across chat and multiplayer.

### Fixed
- Work around a limitation in cockroachdb with type information in group sub-queries.

## [0.11.0] - 2017-02-09
### Added
- Add `--verbose` flag to enable debug logs in server.
- Database name can now be set in migrations and at server startup. i.e. `nakama --db root@127.0.0.1:26257/mydbname`.
- Improve SQL compatibility.

### Changed
- Update db schema to support 64 characters with device IDs. This enables `SystemInfo.deviceUniqueIdentifier` to be used as a source for device IDs on Windows 10.
- Logout messages now close the server-side connection and won't reply.
- Rename logout protocol message type from `TLogout` to `Logout`.
- Update server protocol for friend messages to use IDs as bytes.

### Fixed
- Fix issue where random handle generator wasn't seeded properly.
- Improve various SQL storage, friend, and group queries.
- Send close frame message in the websocket to gracefully close a client connection.
- Build system will now detect modifications to `migrations/...` files and run dependent rules.

## [0.10.0] - 2017-01-14
### Added
- Initial public release.

