// Package unofficial provides constants for well-known HTTP status codes that
// are not part of the official specification.
package unofficial

// StatusThisIsFine is used by Apache servers as a catch-all error condition
// allowing the passage of message bodies through the server when the
// ProxyErrorOverride setting is enabled.
const StatusThisIsFine = 218

// StatusPageExpired is used by the Laravel Framework when a CSRF Token is
// missing or expired.
const StatusPageExpired = 419

// StatusMethodFailure is used by the Spring Framework when a method has failed.
const StatusMethodFailure = 420

// StatusEnhanceYourCalm is used by version 1 of the Twitter Search and Trends
// API when the client is being rate limited.
const StatusEnhanceYourCalm = 420

// StatusRequestHeaderFieldsTooLarge is used by Shopify, instead of the 429
// response code, when too many URLs are requested within a certain time frame.
const StatusRequestHeaderFieldsTooLarge = 430

// StatusLoginTimeout is used by IIS when the client's session has expired and
// must log in again.
const StatusLoginTimeout = 440

// StatusNoResponse is used by NGINX internally to instruct the server to return
// no information to the client and close the connection immediately.
const StatusNoResponse = 444

// StatusRetryWith is used by IIS when the server cannot honour the request
// because the user has not provided the required information.
const StatusRetryWith = 449

// StatusBlockedByWindowsParentalControls is used by Windows Parental Controls
// when blocking access to the requested webpage.
const StatusBlockedByWindowsParentalControls = 450

// StatusRedirect is used by Exchange ActiveSync when either a more efficient
// server is available or the server cannot access the users' mailbox.
const StatusRedirect = 451

// StatusClientClosedConnection is used by AWS Elastic Load Balancing when the
// client closed the connection with the load balancer before the idle timeout
// period elapsed.
const StatusClientClosedConnection = 460

// StatusXForwardedForTooLarge is used by AWS Elastic Load Balancing when the
// load balancer received an X-Forwarded-For request header with more than 30
// IP addresses.
const StatusXForwardedForTooLarge = 463

// StatusIncompatibleProtocolVersions is used by AWS Elastic Load Balancing
// when the client and origin server are using incompatible protocol versions.
const StatusIncompatibleProtocolVersions = 464

// StatusRequestHeaderTooLarge is used by NGINX when the client sent a request
// or header that was too large.
const StatusRequestHeaderTooLarge = 494

// StatusSSLCertificateError is used by NGINX when the client has provided an
// invalid client certificate.
const StatusSSLCertificateError = 495

// StatusSSLCertificateRequired is used by NGINX when a client certificate is
// required but not provided.
const StatusSSLCertificateRequired = 496

// StatusHTTPRequestSentToHTTPSPort is used by NGINX when the client has made an
// HTTP request to a port listening for HTTPS requests.
const StatusHTTPRequestSentToHTTPSPort = 497

// StatusInvalidToken is used by ArcGIS for Server when a token is expired or
// otherwise invalid.
const StatusInvalidToken = 498

// StatusTokenRequired is used by ArcGIS for Server when a token is required
// but was not submitted.
const StatusTokenRequired = 499

// StatusClientClosedRequest is used by NGINX when the client has closed the
// request before the server could send a response.
const StatusClientClosedRequest = 499

// StatusBandwidthLimitExceeded is used by Apache servers and cPanel when the
// server has exceeded the bandwidth specified by the server administrator.
const StatusBandwidthLimitExceeded = 509

// StatusWebServerReturnedAnUnknownError is used by Cloudflare when the origin
// server returned an empty, unknown, or unexpected response.
const StatusWebServerReturnedAnUnknownError = 520

// StatusWebServerIsDown is used by Cloudflare when the origin server refused
// the connection.
const StatusWebServerIsDown = 521

// StatusConnectionTimedOut is used by Cloudflare when the connection timed out
// contacting the origin server.
const StatusConnectionTimedOut = 522

// StatusOriginIsUnreachable is used by Cloudflare when it could not reach the
// origin server.
const StatusOriginIsUnreachable = 523

// StatusTimeoutOccurred is used by Cloudflare when it was able to complete a
// TCP connection to the origin server, but did not receive a timely response.
const StatusTimeoutOccurred = 524

// StatusSSLHandshakeFailed is used by Cloudflare when it could not negotiate a
// SSL/TLS handshake with the origin server.
const StatusSSLHandshakeFailed = 525

// StatusInvalidSSLCertificate is used by Cloudflare when it could not validate
// the SSL certificate of the origin web server.
const StatusInvalidSSLCertificate = 526

// StatusRailgunError is used by Cloudflare when the connection to the origin
// server's Railgun server is interrupted.
const StatusRailgunError = 527

// StatusSiteIsOverloaded is used by Qualys in the SSLLabs server testing API to
// signal that the site can not process the request.
const StatusSiteIsOverloaded = 529

// StatusSiteIsFrozen is used by the Pantheon Systems web platform to indicate a
// site that has been frozen due to inactivity.
const StatusSiteIsFrozen = 530

// StatusCloudflareError is used by Cloudflare when returning a 1xxx error.
const StatusCloudflareError = 530

// StatusUnauthorized is used by AWS Elastic Load Balancing when the identity
// provider returned an error code when authenticating the user.
const StatusUnauthorized = 561

// StatusNetworkReadTimeout is used by some HTTP proxies to signal a network
// read timeout behind the proxy to a client in front of the proxy.
const StatusNetworkReadTimeout = 598

// StatusNetworkConnectTimeout is used by some HTTP proxies to signal a network
// connect timeout behind the proxy to a client in front of the proxy.
const StatusNetworkConnectTimeout = 599
