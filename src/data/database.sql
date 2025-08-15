CREATE TABLE IF NOT EXISTS "users" (
  "phone_number" TEXT NOT NULL UNIQUE,
  "name" TEXT NOT NULL,
  PRIMARY KEY("phone_number")
);

CREATE TABLE IF NOT EXISTS "messages" (
  "id" INTEGER NOT NULL UNIQUE,
  "sender" TEXT NOT NULL,
  "receiver" TEXT NOT NULL,
  "content" TEXT NOT NULL,
  "timestamp" TEXT NOT NULL,
  PRIMARY KEY("id" AUTOINCREMENT),
  FOREIGN KEY("sender") REFERENCES users("phone_number"),
  FOREIGN KEY("receiver") REFERENCES users("phone_number")
);
