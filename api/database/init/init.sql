CREATE TABLE user (
    id char(26) NOT NULL PRIMARY KEY,
    name varchar(50) NOT NULL,
    age int(3) NOT NULL
);

INSERT INTO user VALUES ('00000000000000000000000001', 'hanako', 20);
INSERT INTO user VALUES ('00000000000000000000000002', 'taro', 30);

CREATE TABLE posts (
    id char(26) NOT NULL PRIMARY KEY,
    userid varchar(50) NOT NULL,
    name varchar(50) NOT NULL,
    times varchar(50) NOT NULL,
    likes int(5) NOT NULL,
    retweet int(4) NOT NULL,
    content varchar(300) NOT NULL,
    reply_to char(26)
);
