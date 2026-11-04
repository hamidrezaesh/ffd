package cmd

var commandsHelp string = `usage: ffd [URL]...[OPTION]

Commands:
proxy              Start the ffd forward proxy
Example: ffd proxy

update             Update ffd to the latest version
Example: ffd update

Startup:
-h, --help	Show help

-v, --version	Show ffd version

Options:
-o, --output NAME	Save the file with a custom filename
Example: ffd <URL> -o my-file

-w, --wait SECONDS	Wait before starting the download
Example: ffd <URL> -w 100

-p, --path PATH	save the file to a custom directory (default .)
Example: ffd <URL> -p /path/to/your/folder

-r --max-retries NUMBER_OF_RETRIES	Total retries after connection failed (default 4)
Example: ffd <URL> -r 10

-W --max-workers NUMBER_OF_WORKERS	Total concurrent workers (default 8)
Example: ffd <URL> -W 10

-c --max-chunks NUMBER_OF_CHUNKS	Total parts of download (default 12)
Example: ffd <URL> -c 20

--protocol PROTOCOL	Protocol to use for download (default 'auto')
Example: ffd <URL> --protocol http2

--set-proxy PROXY    Set proxy server for download
Example: ffd <URL> --set-proxy your-proxy

--header HEADER		Set custom header
Example: ffd <URL> --header "User-Agent: Mozilla/5.0"

--cookie COOKIE		Set custom cookie (default "")
Example: ffd <URL> --cookie "session=abc123"

--cookie-file		Load cookie from file
Example: ffd <URL> --cookie-file cookies.txt

--overwrite   Overwrite an existing file and restart the download
Example: ffd <URL> --overwrite
`

var proxyHelp string = `usage: ffd proxy [OPTION]

Startup:
-h, --help             Show help

Network:
--port PORT            Port to run the proxy on (default 8000)
--set-proxy PROXY      Set upstream proxy for FFD

Download:
--max-retries N        Maximum download retries (default 4)
--max-workers N        Maximum concurrent workers (default 8)
--max-chunks N         Maximum download chunks (default 12)
--protocol PROTOCOL    Preferred protocol: auto, http1, http2, http3

Request:
-H, --header HEADER    Add an upstream HTTP header
Example: ffd proxy --header "Authorization: Bearer TOKEN"

Examples:
ffd proxy --port 9000
ffd proxy --set-proxy http://127.0.0.1:8080
ffd proxy --max-workers 16 --max-chunks 24
ffd proxy --protocol http3
ffd proxy --header "User-Agent: ffd/proxy"
`
