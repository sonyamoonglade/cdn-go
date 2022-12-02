# Documentation and overview.
[![forthebadge made-with-go](http://ForTheBadge.com/images/badges/made-with-go.svg)](https://go.dev/)

![MongoDB](https://img.shields.io/badge/MongoDB-%234ea94b.svg?style=for-the-badge&logo=mongodb&logoColor=white)

## Features:
1. File server
2. File processor
3. Module-resolver technique
4. JWT Authorization


# Navigation
### - [Modules and Resolvers](#modules-and-resolvers)
### - [Dealing with Modules as client](#dealing-with-modules-and-resolvers-as-a-client)
### - [Uploading a File](#uploading-a-file)
### - [Getting a File](#getting-a-file)
### - [Deleting a File](#deleting-a-file)
### - [Operations and Security](#operations-and-security)
### - [How to access private resource](#dealing-with-tokens)

--- 

## Modules and resolvers

CDN has terms "**module**" and "**resolver**".

### What is a Module?

> **Module**- *unique set of resolvers grouped by mime-type*
 
### What is a Resolver?

> **Resolver** - *function or function name, responsible for file processing*

### Example: 
>"*Image*" **Module** has **Resolver** called "crop".
> 
> Logically, this module is responsible for processing images.
> 
> And functionality of "crop" is to crop images.

## Must-to know

*CDN offers Bucket-Like storage.\
Buckets are specified and created according to mime-type.\
Therefore, only **one** module can be assigned to a bucket.*

### Example:
> When bucket is created, module is assigned to it
> 
> Assuming we have a bucket called "static-site" and it has **Module** "*image*".
> 
> This module can have many resolvers. They could be called in a single chain.

---

# Dealing with modules and resolvers (as a client)
Modules and resolvers are sent via URL Query.\
For example, client wants to get cropped image file using "*image*" **Module** and "*crop*" **Resolver**.

*First part of url is omitted*

	{module}.{resolver}={arg} // example

	GET ...?image.crop=true

Take a look at how combination of **Module** + **Resolver** + **Arg** is passed.

Usually **Resolvers** don't come alone. \
They have arguments (**Arg**)

Most of the resolvers are simply offering 'true' or 'false' arguments. \

Occasionally, some resolvers might have custom and specific arguments.\
For example:
- normal
- big
- large
- 80

etc...

### Possible errors
> **MRA** - Module Resolver Argument

If client requests a file from bucket "site-content" and sends multiple **MRA's**. 

For example:

	GET ...?image.crop=true&music.compress=true

*image* **Module** and *music* **Module** are used in one request.\
CDN will respond with 400 *BAD REQUEST* code.

As stated, **only one module can be used for 1 bucket!**

--- 

# Uploading a File

Upload to CDN is done via FormData.\
CDN can handle bulk upload (FormData with > 1 file).

Uploading flow:  **Client** --> **Backend** --> **CDN**\
Backend knows what bucket an incoming file should be uploaded to.

>Assume we are uploading to a bucket called "site-content"
	
	http(s)://cdn.domain.com/{bucket} // example

	POST http(s)://cdn.domain.com/site-content
	+ FormData

CDN will save files to a `{bucket}` ("site-content") and return two [**String Array**] lists: list of IDs assigned to uploaded files and list of URLs which files can be accessed with.

### Example return:
	201 OK

	{
	   "ids": ["1234-abcd-4567-fghk"],
	   "urls": ["cdn.domain.com/site-content/1234-abcd-4567-fghk"]
	}



---

# Getting a file

When file is uploaded, see [Uploading a File](#uploading-a-file), it's given an ID.\
Knowing the bucket you can get the file by running the following request

	http(s)://cdn.domain.com/{bucket}/{id} // example

	GET http(s)://cdn.domain.com/site-content/1234-abcd-4567-fghk

This will return original file and 200 *OK*.

### Using Modules and Resolvers

> Assume "*image*" **Module** and **Resolver** "*lowres*" +  **Arg** *"true"* are documented.

If you need to load file processed by a module.\
For example, on initial page load client wants to load low-resolution image.

For this purpose you should send the following request

	GET http(s)://cdn.domain.com/site-content/1234-abcd-4567-fghk?image.lowres=true

This will return processed low resolution (webp) file and 200 *OK*

[Possible errors](#possible-errors)

# Deleting a file

At this point you must already know file's ID and bucket.\
For our example we'd take those parameters from upper examples.

	http(s)://cdn.domain.com/{bucket}/{id} // example
	
	DELETE http(s)://cdn.domain.com/site-content/1234-abcd-4567-fghk

This will return 200 *OK* if resource is successfully deleted.

[Possible errors](#possible-errors)

# Operations and security

**CDN offers JWT Authorization as security**.

### What is an operation?
> **Operation** - *certain type of API-Call (request) to the CDN*

There are **3** operations for buckets:
1. get
2. post
3. delete

## Authentication
Each **Operation** from the list has type **Public** or **Private**.

>**Private operation** - *an operation that requires JWT signed token passed with request*.
>
>**Important note** - only one resource (file) could be accessed with one jwt token. 
> This is an option because each JWT token has it's own fileID inside a payload. See [Payload](#Payload)

>**Public operation** - *an operation that has public API for everyone*

To sign JWT tokens CDN uses *Keys*.



### What is a key?
> **Key** - *private string which JWT token is signed with*

Each operation, when bucket is created, can have multiple keys.

Let's assume our bucket "site-content" has the following config:
	
	get: {
	  type: public,
	  keys: []
	},
	post: {
	  type: private,
 	  keys: ["example.xyz", "xyz.example"]
	},
	delete: {
	  type: private,
	  keys: ["xxx.example.xxx]
	}

As we can see, *get* operation is public, so [Get file](#getting-a-file) will be public to everybody.
### Important to know
For *post* operation for example, if JWT token is signed with `"example.xyz"` secret - **it's valid**.
Moreover, if JWT is signed with `"xyz.example""` then token is also **valid**

### Dealing with tokens
Speaking of *delete* and *post*, they have defined keys.\
To make an API-Call (to upload file or to delete the resource) the sender\
must sign JWT token with `one of the keys` secret and pass it via `Authorization` header.

Every single JWT token must have payload in it.

### Payload
	
	
	{
	  "bucket": "{requested_bucket}", // same as for http request
	  "file_id": "{requested_file_uuid}" // same as for http request
	} // example

	{
	  "bucket": "site-content",
	  "file_id": "1234-abcd-4567-fghk"
	}

- You can't access a file (file_id=1234) if your JWT token has payload (file_id=5678).
- You can't access a file inside (bucket=images) if your JWT token has payload (bucket=site-content).


CDN would return:

	StatusCode: 403 Forbidden

	{
		"message": "access denied"
	}

---

**Upload**

	http(s)://cdn.domain.com/{bucket} // example

	POST http(s)://cdn.domain.com/site-content
	+ FormData with file
	
	headers: {
	  Authorization: Bearer {your_token},
	  ...
	}

---

**Delete**
	
	http(s)://cdn.domain.com/{bucket}/{id} // example

	DELETE http(s)://cdn.domain.com/site-content/abcd-1234-defg
	
	headers: {
	  Authorization: Bearer {your_token},
	  ...
	}

---

**Get**
	
	http(s)://cdn.domain.com/{bucket}/{id}?auth={your_token}
	
	GET http(s)://cdn.domain.com/site-content/abcd-1234-defg?auth=exampletoken

---

Pay attention to how token is passed in different cases.




### Possible errors:
- Missing token -> 401 Unauthorized
- Invalid format, Invalid token payload, Invalid signature-> 403 Forbidden


# Currently supported and implemented modules
This part of documentation is essential and will provide comprehensive and detailed\
information about all modules, resolvers and arguments that's currently implemented.

#### Template:

### [Module]
- **Resolver 1** - *some example functionality...*
  - arg1: *effects...*
  - arg2: *extra effects...*



## Image
- **webp** - *Makes image to .webp format. It significantly reduces image size\
without quality loss. For example: **Initial image - 4MB,** and it's .webp clone - **600KB***
	- true: if passed, cdn would process image to .webp format and return it.
    - false: default. If passed nothing would happen.
  

- **resized** - *Makes image suitable for wide devices. Reduces size\
from the top and bottom by 21.14%, thus image becomes less\
at height and can fit on wider screen*
  - **true**: if passed, cdn would crop image to ~ 60% of initial size and return it.
  - **false**: default. If passed nothing would happen.


 **Important** - size is reduced **both** from the top and bottom so image perspective stays the same.


# Run in docker
Save the file to trigger hot reload 

	dev environment including hot reload

	$ make run 

---

	stop mongodb and cdn containers
	
	$ make stop




	



