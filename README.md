# Site Monitor using goengine

This application is a Site Monitor written in Go that is based on [GoEngine](https://github.com/ninnemana/goengine).

## Getting Started

db.sql is the database generation script to get you started.
You'll need to create a file called ConnectionString.go in the helpers/database folder. It should look like this:

    package database

    const (
        db_proto = "tcp"
        db_addr  = "dbserveraddress:3306"
        db_user  = "dbusername"
        db_pass  = "dbpassword"
        db_name  = "dbname"
    )

After you've created the database, set up permissions, run the generation script and created this file, you should be good to start using the application. You'll find the admin located at /admin.

## Credits

Written By: [Jessica Janiuk](https://github.com/janiukjf)
Special Thanks: [Alex Ninneman](https://github.com/ninnemana)

## TO DO

* Add Internal Auth
* Add User management