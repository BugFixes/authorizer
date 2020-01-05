-- DROP TABLE "public"."agent";

CREATE TABLE "public"."agent" (
                                  "id"         uuid,
                                  "name"       varchar(200),
                                  "key"        uuid,
                                  "secret"     uuid,
                                  "company_id" uuid,
                                  PRIMARY KEY ("id")
);
--
-- CREATE TABLE "public"."company" (
--                                     "id" uuid,
--                                     "name" varchar(200),
--                                     PRIMARY KEY ("id")
-- );
--
-- ALTER TABLE "agent" ADD FOREIGN KEY ("companyId") REFERENCES "company" ("id");
