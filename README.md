# Mastering Go Web Services

## Setup MySQL

Start MySQL console as root: `mysql -u root` and create database:

```
create database social_network;
``` 

Then create database user and grant access for this user to the newly
created database:

```
grant
```

Create `users` table:

```
CREATE TABLE users (
    user_id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    user_nickname VARCHAR(32) NOT NULL,
    user_first VARCHAR(32) NOT NULL,
    user_last VARCHAR(32) NOT NULL,
    user_email VARCHAR(128) NOT NULL,
    PRIMARY KEY (user_id),
    UNIQUE INDEX user_nickname (user_nickname)
);
```

Add new unique index to `users` table:

```
ALTER TABLE users ADD UNIQUE INDEX user_email (user_email); 
```

Create new `users_relationships` table:

```
CREATE TABLE users_relationships (
    users_relationship_id INT(13) NOT NULL AUTO_INCREMEMT,
    from_user_id INT(10) NOT NULL,
    to_user_id INT(10) UNSIGNED NOT NULL,
    users_relationship_type VARCHAR(10) NOT NULL,
    users_relationship_timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (users_relationship_id),
    INDEX from_user_id (from_user_id),
    INDEX to_user_id (to_user_id),
    INDEX from_user_id_to_user_id (from_user_id, to_user_id),
    INDEX from_user_id_to_user_id_users_relationship_type (from_user_id, to_user_id, users_relationship_type)
);
```

## Addition for Chapter 3
Add new field to `users` table:

```
ALTER TABLE users
  ADD COLUMN user_image MEDIUMBLOB NOT NULL AFTER user_email;
```
