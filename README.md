# 62dolphin
  62dolphin microservice is a REST API writter in Golang designed for authenticaction with Json Web Token

  Created by 62teknologi.com, perfected by Community.

<details>
<summary><b>View table of contents</b></summary>

- [authentication](#authenticaction)
  - [concept](#concept)
    - [Authentication Information](#authentication-information)
    - [Authentication Behaviours](#authentication-ehaviors)
    - [Authentication Associations](#authentication-associations)
  - [Running 62dolphin](#running-62dolphin)
    - [Prerequisites](#prerequisites)
    - [Installation manual](#installation-manual)
  - [API Endpoints](#api-endpoints)
-[contributing](#contributing)
  - [Must Preserve Characteristic](#must-preserve-characteristic)
  - [License](#license)
- [About](#about-62)
- [Why you should use this payment proxy?](#why-you-should-use-this-payment-proxy)
- [Current Limitations](#current-limitations)
- [Implemented Channels](#implemented-channels)
- [Getting Started](#getting-started)
  - [Payment Gateway Registration](#payment-gateway-registration)
    - [Midtrans](#midtrans)
    - [Xendit](#xendit)
    - [Midtrans VS Xendit Onboarding](#midtrans-vs-xendit-onboarding)
  - [Payment Gateway Callback](#payment-gateway-callback)
    - [Midtrans](#midtrans-1)
    - [Xendit](#xendit-1)
  - [Application Secret](#application-secret)
    - [Database](#database)
    - [Midtrans Credential](#midtrans-credential)
    - [Xendit Credential](#xendit-credential)
  - [Configuration File](#configuration-file)
  - [Mandatory Environment Variables](#mandatory-environment-variables)
- [Example Code](#example-code)
- [API Usage](#api-usage)
- [Contributing](#contributing)
- [License](#license)

</details>

## Authentication
This introduction will help You explain the concept and characteristic of Authenetication.

### Concept
Authentication is the process of verifying the identity of a user or other entity in a system. It is done to ensure that only authorized users or entities can access resources or perform certain actions.

In the context of 62dolphin, authentication refers to the process of verifying user credentials (such as username and password) to grant access to protected resources or services. This is typically done using an authentication token, such as a Json Web Token (JWT).

### Authentication Information
- User Credentials (e.g., username and password)
- Authentication Token (e.g., JWT)
- Token Expiration Time
- Access Rights or Permissions

### Authentication Behaviors
- Can Authenticate (Log in)
- Can Verify Authentication Token
- Can Refresh Authentication Token
- Can Revoke Access or Invalidate Authentication Token

### Authentication Associations
- Associated with the Authenticated User or Entity
- Associated with the Protected Resources or Services
- Associated with the Granted Access Rights or Permissions



## Running 62dolphin

Follow the instruction below to running 62dolphin on Your local machine.

### Prerequisites
Make sure to have preinstalled this prerequisites apps before You continue to installation manual. we don't include how to install these apps below most of this prerequisites is a free apps which You can find the "How to" installation tutorial anywhere in web and different machine OS have different way to install.
- MySql
- Go

### Installation manual
This installation manual will guide You to running the binary on Your ubuntu or mac terminal.

1. Clone the repository
```
git clone https://github.com/62teknologi/62dolphin.git
```

2. Change directory to the cloned repository
```
cd 62dolphin
```

3. Initiate the submodule
```
git submodule update --init
```

3. tidy
```
go mod tidy
```

4. Create .env base on .env.example
```
cp .env.example .env
```

5. setup env file like this
```
ENVIRONMENT=development
DB_SOURCE_1=user:db_pass@tcp(db_ip:db_port)/db_name?charset=utf8mb4&parseTime=True&loc=Local
DB_SOURCE_2=
HTTP_SERVER_ADDRESS=0.0.0.0:7001
MONOLITH_URL=http://localhost:8001
API_KEY=

FIREBASE_CONFIG_PATH=

TOKEN_SYMMETRIC_KEY=12345678901234567890123456789012
ACCESS_TOKEN_DURATION=24h
REFRESH_TOKEN_DURATION=87h
```

6. Build the binary
```
go build -v -o 62dolphin main.go
```

7. Run the server
```
./62dolphin
```

The API server will start running on `http://localhost:7001`. You can now interact with the API using Your preferred API client or through the command line with `curl`.

## API Endpoints

#### Endpoint
```
GET    /health                   
POST   /api/v1/auth/sign-in      
POST   /api/v1/auth/sign-up     
GET    /api/v1/auth/:adapter     
GET    /api/v1/auth/:adapter/callback 
POST   /api/v1/auth/:adapter/callback 
POST   /api/v1/auth/:adapter/verify 
POST   /api/v1/otps/create       
POST   /api/v1/tokens/create     
POST   /api/v1/tokens/verify     
POST   /api/v1/tokens/refresh    
POST   /api/v1/passwords/create  
POST   /api/v1/passwords/check   
POST   /api/v1/passwords/forgot  
PATCH  /api/v1/passwords/reset/:token 
GET    /api/v1/users             
POST   /api/v1/users             
GET    /api/v1/users/:id         
POST   /api/v1/users/verify      
POST   /api/v1/tokens/block      
POST   /api/v1/tokens/block-all  
PUT    /api/v1/users/:id         
DELETE /api/v1/users/:id         
```

You can check all endpoint on [docs](/docs/62Dolphin-microservice.postman_collection.json)

# Contributing

If You'd like to contribute to the development of the 62whale REST API, please follow these steps:

1. Fork the repository
2. Create a new branch for Your feature or bugfix
3. Commit Your changes to the branch
4. Create a pull request, describing the changes You've made

We appreciate Your contributions and will review Your pull request as soon as possible.

## Must Preserve Characteristic 
- Reduce repetition
- Easy to use REST API
- Easy to setup
- Easy to Customizable
- high performance
- Robust data validation and error handling
- Well documented API endpoints

## License

This project is licensed under the MIT License. For more information, please see the [LICENSE](./LICENSE) file.

# About 62
**E.nam\Du.a**

Indonesian language; spelling: A-num\Due-wa

Origin: Enam Dua means ‘six-two’ or sixty two. It is Indonesia’s international country code (+62), that was also used as a meme word for “Indonesia” by “Indonesian internet citizen” (netizen) in social media.