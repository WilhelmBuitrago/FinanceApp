package domain

import "time"

/* NowRFC3339 is the one way the app stamps a row. The codebase carried
   eighteen inline copies of the same expression and one local helper in
   the api package that services could not see; every writer now names
   the same moment the same way. */

func NowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}
