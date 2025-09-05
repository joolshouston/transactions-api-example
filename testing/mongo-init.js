db.getSiblingDB('pismo')

db.createUser(
        {
            user: "testuser",
            pwd: "password",
            roles: [
                {
                    role: "readWrite",
                    db: "pismo"
                }
            ]
        }
);

db.createCollection('accounts');