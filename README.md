Gator - RSS Feed Reader
Welcome to Gator!
This is a RSS feed reader cli.

Requirements:
    go
    postgres 
Install
Step 1. create a config file in your home directory called .gatorconfig.json
.gatorconfig.json should have this structure
{
{"db_url":"postgres://example_username:example_password@localhost:5432/gator?sslmode=disable","current_user_name":"user"}
}

replace example_username and example_password

Step 2. build the gator binary in the main folder.
go build

Step 3.

run
gator migrate up

Gator is now ready to use. run gator help to get an overview of the possible commands.



