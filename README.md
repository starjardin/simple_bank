### How to generate code

- Generate SQL CRUD with sqlc:

    ```bash
    make sqlc
    ```

- Generate DB mock with gomock:

    ```
    make mock
    ```

- Create a new db migration 

    ```bash
    migrate create -ext sql -dir db/migration -seq <migration-name>
    ```

### How to run

- Run server:

    ```bash
    make server
    ```