// Copyright 2019 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ngconf

import (
	"fmt"
)

func ExampleDecode() {
	const conf = `
    user www-data;
    worker_processes auto;
    pid /run/nginx.pid;
    include /etc/nginx/modules-enabled/*.conf;

    events {
            worker_connections 768;
            # multi_accept on;
    }

    stream {
        server {
            listen 127.0.0.1:8443;
            proxy_connect_timeout 1s;
            proxy_pass backend;
        }
    }


    http {

            ##
            # Basic Settings
            ##

            sendfile on;
            tcp_nopush on;
            tcp_nodelay on;
            keepalive_timeout 65;
            types_hash_max_size 2048;
            # server_tokens off;

            # server_names_hash_bucket_size 64;
            # server_name_in_redirect off;

            include /etc/nginx/mime.types;
            default_type application/octet-stream;

            ##
            # SSL Settings
            ##

            # Dropping SSLv3, ref: POODLE
            ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
            ssl_prefer_server_ciphers on;

            ##
            # Logging Settings
            ##

            access_log /var/log/nginx/access.log;
            error_log /var/log/nginx/error.log;

            ##
            # Gzip Settings
            ##

            gzip on;

            # gzip_vary on;
            # gzip_proxied any;
            # gzip_comp_level 6;
            # gzip_buffers 16 8k;
            # gzip_http_version 1.1;
            # gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;

            ##
            # Virtual Host Configs
            ##

            include /etc/nginx/conf.d/*.conf;
            include /etc/nginx/sites-enabled/*;
    }

    #mail {
    #       # See sample authentication script at:
    #       # http://wiki.nginx.org/ImapAuthenticateWithApachePhpScript
    #
    #       # auth_http localhost/auth.php;
    #       # pop3_capabilities "TOP" "USER";
    #       # imap_capabilities "IMAP4rev1" "UIDPLUS";
    #
    #       server {
    #               listen     localhost:110;
    #               protocol   pop3;
    #               proxy      on;
    #       }
    #
    #       server {
    #               listen     localhost:143;
    #               protocol   imap;
    #               proxy      on;
    #       }
    #}`

	root, err := Decode(conf)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Remove a config block.
	root.Del("events")

	// Add a new config item.
	stream := root.Get("stream")[0]
	backend := stream.Add("upstream", "backend")
	backend.Add("hash", "$remote_addr", "consistent")
	backend.Add("server", "backend1:443", "max_fails=3", "fail_timeout=30s")

	// // Get a config item.
	backend = stream.Get("upstream", "backend")[0]
	hash := backend.Get("hash", "$remote_addr")
	fmt.Println(len(hash), "node:", hash[0].Directive, hash[0].Args)

	// Print the whole config.
	fmt.Println("========================= Config =========================")
	fmt.Println(root)

	// Output:
	// 1 node: hash [$remote_addr consistent]
	// ========================= Config =========================
	// user www-data;
	// worker_processes auto;
	// pid /run/nginx.pid;
	// include /etc/nginx/modules-enabled/*.conf;
	//
	// stream {
	//     server {
	//         listen 127.0.0.1:8443;
	//         proxy_connect_timeout 1s;
	//         proxy_pass backend;
	//     }
	//
	//     upstream backend {
	//         hash $remote_addr consistent;
	//         server backend1:443 max_fails=3 fail_timeout=30s;
	//     }
	// }
	//
	// http {
	//     ##
	//     # Basic Settings
	//     ##
	//     sendfile on;
	//     tcp_nopush on;
	//     tcp_nodelay on;
	//     keepalive_timeout 65;
	//     types_hash_max_size 2048;
	//
	//     # server_tokens off;
	//     # server_names_hash_bucket_size 64;
	//     # server_name_in_redirect off;
	//     include /etc/nginx/mime.types;
	//     default_type application/octet-stream;
	//
	//     ##
	//     # SSL Settings
	//     ##
	//     # Dropping SSLv3, ref: POODLE
	//     ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
	//     ssl_prefer_server_ciphers on;
	//
	//     ##
	//     # Logging Settings
	//     ##
	//     access_log /var/log/nginx/access.log;
	//     error_log /var/log/nginx/error.log;
	//
	//     ##
	//     # Gzip Settings
	//     ##
	//     gzip on;
	//
	//     # gzip_vary on;
	//     # gzip_proxied any;
	//     # gzip_comp_level 6;
	//     # gzip_buffers 16 8k;
	//     # gzip_http_version 1.1;
	//     # gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;
	//     ##
	//     # Virtual Host Configs
	//     ##
	//     include /etc/nginx/conf.d/*.conf;
	//     include /etc/nginx/sites-enabled/*;
	// }
	//
	// #mail {
	// #       # See sample authentication script at:
	// #       # http://wiki.nginx.org/ImapAuthenticateWithApachePhpScript
	// #
	// #       # auth_http localhost/auth.php;
	// #       # pop3_capabilities "TOP" "USER";
	// #       # imap_capabilities "IMAP4rev1" "UIDPLUS";
	// #
	// #       server {
	// #               listen     localhost:110;
	// #               protocol   pop3;
	// #               proxy      on;
	// #       }
	// #
	// #       server {
	// #               listen     localhost:143;
	// #               protocol   imap;
	// #               proxy      on;
	// #       }
	// #}
}
