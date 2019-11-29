# nginx

Can generate the `.htpasswd` file with `htpasswd -c /etc/nginx/.htpasswd
username`. We could also potentially specify `password` with `-b`. We can
decide.

Could potentially use the `-B` for bcrypt.

Not entirely sure if its safe to commit the actual passwd file to source
control.
