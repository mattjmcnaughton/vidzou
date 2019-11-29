FROM ubuntu:18.04

RUN apt update
RUN apt install -y nginx apache2-utils

# Obviously insecure because we are commiting the password to source control
# (which I'm not crazy about, even though `.htpasswd` is technically encrypted).
COPY insecure-htpasswd /etc/nginx/.htpasswd
COPY default-site /etc/nginx/sites-enabled/default

# Running as root? Is that bad?
CMD ["nginx", "-g", "daemon off;"]
