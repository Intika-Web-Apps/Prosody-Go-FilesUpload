# Prosody Filer

A simple file server for handling XMPP http_upload requests. This server is meat to be used with the Prosody [mod_http_upload_external](https://modules.prosody.im/mod_http_upload_external.html) module.

*(This module can also be used with future versions of Ejabberd: https://github.com/processone/ejabberd/commit/fface33d54f24c777dbec96fda6bd00e665327fe)*

## Why should I use this server?

* Prosody developers recommend using http_upload_external instead of http_upload (Matthew Wild on the question if http_upload is memory leaking):
    > "BTW, I am not aware of any memory leaks in the HTTP upload code. However it is known to be very inefficient.
    > That's why it has a very low upload limit, and **we encourage people to use mod_http_upload_external instead**.
    > We set out to write a good XMPP server, not  HTTP server (of which many good ones already exist), so our HTTP server is optimised for small bits of data, like BOSH and websocket.
    > Handling large uploads and downloads was not a goal (and implementing a great HTTP server is not a high priority for the project compared to other things).
    > **Our HTTP code buffers the entire upload into memory.
    > More, it does it in an inefficient way that can use up to 4x the actual size of the data (if the data is large).
    > So uploading a 10MB file can in theory use 40MB RAM.**
    > But it's not a leak, the RAM is later cleared and reused. [...]
    > The GC will free the memory at some point, but the OS may still report that Prosody is using that memory due to the way the libc allocator works.
    > Most long lived processes behave this way (only increasing RAM, rarely decreasing)."
* This server works without any script interpreters or additional dependencies. It is delivered as a binary.
* Go is very good at serving HTTP requests.


## Download

If you are using regular x86_64 Linux, you can download a finished binary for your system on the [release page](https://github.com/ThomasLeister/prosody-filer/releases). **No need to compile this application yourself**.


## Build (optional)

If you're using something different than a x64 Linux, you need to compile this application yourself.

To compile the server, you need a full Golang development environment. This can be set up quickly: https://golang.org/doc/install#install

Then checkout this repo:

    go get github.com/ThomasLeister/prosody-filer

and switch to the new directory:

    cd $GOPATH/src/github.com/ThomasLeister/prosody-filer

The application can now be build:

    ### Build static binary
    ./build.sh

    ### OR regular Go build
    go build main.go


## Set up / configuration


### Setup Prosody Filer environment

Create a new user for Prosody Filer to run as:

    adduser --disabled-login --disabled-password prosody-filer

Switch to the new user:

    su - prosody-filer

Copy

* the binary ```prosody-filer``` and
* config ```config.example.toml```

to ```/home/prosody-filer/```. Rename the configuration to ```config.toml```.


### Configure Prosody

Back in your root shell make sure ```mod_http_upload``` is **dis**abled and ```mod_http_upload_external``` is **en**abled! Then configure the external upload module:

```
http_upload_external_base_url = "https://uploads.myserver.tld/upload/"
http_upload_external_secret = "mysecret"
http_upload_external_file_size_limit = 50000000 -- 50 MB
```

Restart Prosody when you are finished:

    systemctl restart prosody

### Configure Prosody Filer

Prosody Filer configuration is done via the config.toml file in TOML syntax. There's not much to be configured:

```
### IP address and port to listen to, e.g. "127.0.0.1:5050"
listenport      = "127.0.0.1:5050"

### Secret (must match the one in prosody.conf.lua!)
secret          = "mysecret"

### Where to store the uploaded files
storeDir        = "./upload/"

### Subdirectory for HTTP upload / download requests (usually "upload/")
uploadSubDir    = "upload/"
```

Make sure ```mysecret``` matches the secret defined in your mod_http_upload_external settings!

In addition to that, the nginx user or group should be able to read the files in
`/home/prosody-filer/upload/`.


### Systemd service file

Create a new Systemd service file: ```/etc/systemd/system/prosody-filer.service```

    [Unit]
    Description=Prosody file upload server

    [Service]
    Type=simple
    ExecStart=/home/prosody-filer/prosody-filer
    Restart=always
    WorkingDirectory=/home/prosody-filer
    User=prosody-filer
    Group=nginx

    [Install]
    WantedBy=multi-user.target

Reload the service definitions, enable the service and start it:

    systemctl daemon-reload
    systemctl enable prosody-filer
    systemctl start prosody-filer

Done! Prosody Filer is now listening on the specified port and waiting for requests.



### Configure Nginx

Create a new config file ```/etc/nginx/sites-available/uploads.myserver.tld```:

    server {
        listen 80;
        listen [::]:80;
        listen 443 ssl;
        listen [::]:443 ssl;

        server_name uploads.myserver.tld;

        ssl_certificate /etc/letsencrypt/live/uploads.myserver.tld/fullchain.pem;
        ssl_certificate_key /etc/letsencrypt/live/uploads.myserver.tld/privkey.pem;

        location /upload/ {
            root /home/prosody-filer;
            client_max_body_size 51m;
            client_body_buffer_size 51m;
            try_files $uri $uri/ @prosodyfiler;
        }

        location @prosodyfiler {
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_set_header X-Forwarded-Host $host:$server_port;
            proxy_set_header X-Forwarded-Server $host;
            proxy_set_header X-Forwarded-For $remote_addr;
            proxy_pass http://127.0.0.1:5050;
            proxy_buffering off;
        }

Enable the new config:

    ln -s /etc/nginx/sites-available/uploads.myserver.tld /etc/nginx/sites-enabled/

Check Nginx config:

    nginx -t

Reload Nginx:

    systemctl reload nginx



## Automatic purge

Prosody Filer has no immediate knowlegde over all the stored files and the time they were uploaded, since no database exists for that. Also Prosody is not capable to do auto deletion if *mod_http_upload_external* is used. Therefore the suggested way of purging the uploads directory is to execute a purge command via a cron job:

    @daily    find /home/prosody-filer/upload -maxdepth 0 -type d -mtime +28 | xargs rm -rf

This will delete uploads older than 28 days.


## Check if it works

Get the log via

    journalctl -f -u prosody-filer

If your XMPP clients uploads or downloads any file, there should be some log messages on the screen.
