Example usage of the Arbiter package.
along with an example of an alternative naive mutex-based strategies (for comparitive benchmarking).

Example use case:
* Idempotent application of a configuration set to an environment, which is tracked by a unique, always increasing version number (eg version '2' supersedes and replaces version '1', etc).  This version is tracked in a keystore to improve performance (no need to query 'on disk' configuration to obtain currently installed version #).  Updates contain all information needed, so applying version '5' over version '3' does not incur any loss of configuration data (thus configuration application is idempotent).
* If the application of the configuration set to the environment fails, the version being stored in the tracking keystore is not updated, and error returned to caller.  For a deterministic simulation of intermittent failures, if the version number is prime, it fails to update.
* If an older or redundant (equal to current) version update comes in (due to delayed/ duplicate processing upstream), the config update is rejected, and an error returned to the caller.

