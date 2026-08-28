#!/bin/bash

# NOTE: Superseded. The project now runs on PostgreSQL — use
# scripts/setup_postgres.sh instead. This script is kept for reference
# only (original CockroachDB local dev flow).

# before we begin i'ld like to say that we're defaulting to talking to
# cockroach in a secure channel because the --insecure option alows us to 
# send queries without authorization through certificates,
# the username:password@hostname:port won't help because by defualt
# comes with one user which is root, and we're not root nor do 
# we have permission priviledges at zone.
# but we need another already instanciated inside the dbms so when we send queries
# from our backend we're specifying our username and not the default root
# so to work around that we have to create another authentication system
# that uses certificates and keys to recognize and trust clients connecting to our
# databases, the ./cockroach cert create.... command will help us alot on that
# as i've shown below.

# so here is how it's gonna go,  if you execute this script it prepares the 
# directorie to store the certs the create.... command expects 
# the generate ca cert and key then node cert and key, then certs for clients that will connect to us

# now to register ourselves we'd open cockroach interactively to create a db user
# with matching our usernames and our password of choice with this command:
# ./cockroach sql --certs-dir=certs
# 	CREATE USER <replace_with_username> WITH PASSWORD 'your_prefered_db_password';
# 	
# 	# after that we need to create our database
# 	CREATE DATABASE skillmatch;
#
# first make sure you're on the directory contain the cockroachdb executable

# first and foremost we create directories for certs and the key
mkdir certs keys

# first we prepare the main authoritative cert and key as we are the CA(certificate authority)
./cockroach cert create-ca --certs-dir=certs --ca-key=keys/ca.key

# first we then create our node/server cert and key
# we specify any name/ip for clients we want our server to trust if they initiate a conneciton
./cockroach cert create-node $HOSTNAME $(hostname -I) localhost 127.0.0.1 --certs-dir=certs --ca-key=keys/ca.key


# for cockroach reasons we need to first create the certs and keys for the root user
# which comes prein, because the only user configured to run cockroach sql queries interactively is she only
# but once we're inside the interactive cockroach we'll create new usernames
./cockroach cert create-client root --certs-dir=certs --ca-key=keys/ca.key

# first we create then a client certificate and key we recognize
# we also store this in the certs directory
# database authorization will just check this hostname whether there's a matching cert and key
./cockroach cert create-client $(whoami) --certs-dir=certs --ca-key=keys/ca.key

# and with that we can also first query cockroach database management system interactively with this command:
# ./cockroach sql --certs-dir=certs --ca-key=keys/ca.key

printf "\nyou're all set, cockroach says, 'have fun'\n\n"



# additional instructions
# to start the dbms:
# ./cockroach start-single-node --certs-dir=certs
#
# now on our go backend we need to put a database connection url in the .env file
# here is an example that worked for me:
# postgresql://skinyanju@LAP-010:26257/skillmatch?sslcert=../cockroach-v26.2.0.linux-amd64/certs%2Fclient.skinyanju.crt&sslkey=../cockroach-v26.2.0.linux-amd64/certs%2Fclient.skinyanju.key&sslmode=verify-full&sslrootcert=../cockroach-v26.2.0.linux-amd64/certs%2Fca.crt
# now this will only work if you have the cockroach directory in the project folder 
# as well you'll need to configure the username to yours specifically
