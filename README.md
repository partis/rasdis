# rasdis

Rasdis is a service discovery tool for Forum Sentry, it uses the rest api to pull back details of policies configured and presents them in swagger ui.

Requires custom swagger-ui package to be available in same directory as bin file and rasdis.cfg (described below).

Configuration
rasdis.cfg is a json file, it can specify the following;

- ForumURL - the url of the Forum Sentry rest api
- ForumUsername - the username to use to connect to Forum Sentry rest api
- ForumPass - the password to use to connect to the Forum Sentry rest apo

