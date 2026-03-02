### Security Package
***
> consist utilty for perform authentication and authorization
 
#### Utilities provided
- Generating single jwt token with expiration
- Generating jwt refresh token mechanism
- Retreiver user info from Context, it is called UserManifest only contain frequently used data such as; `lang,` `timezone`, `authCode`
- Enable user session store on cache, thus enforcing single-user policies, session also persist data entity to enhance performance toward querying user info