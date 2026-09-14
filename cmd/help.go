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
`

var proxyHelp string = `usage: ffd proxy [OPTION]

Startup:
-h, --help        Show help

Options:
--port PORT   Port to run the proxy on (default 8000)
Example: ffd proxy --port 9000

--set-proxy PROXY    Set upstream proxy for ffd proxy
Example: ffd proxy --set-proxy your-proxy
`
