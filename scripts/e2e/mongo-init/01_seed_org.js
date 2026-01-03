// Seed a single GrowthBook organization so SECRET_API_KEY auth works in single-org mode.
// This runs automatically on fresh mongo volumes (docker-entrypoint-initdb.d).

const dbname = "growthbook";
const orgId = "org_gbgen_e2e";

const db2 = db.getSiblingDB(dbname);

db2.organizations.insertOne({
  id: orgId,
  name: "gbgen e2e",
  ownerEmail: "e2e@example.com",
  dateCreated: new Date(),
  members: [],
  invites: [],
});

// Seed a deterministic feature so gbgen generation has stable non-empty output.
db2.features.insertOne({
  id: "e2e-flag",
  organization: orgId,
  description: "E2E seeded feature",
  valueType: "boolean",
  defaultValue: "false",
  owner: "e2e@example.com",
  dateCreated: new Date(),
  dateUpdated: new Date(),
  version: 1,
  environmentSettings: {
    production: {
      enabled: true,
      rules: [],
    },
  },
});

// Seed a corresponding published revision so the API returns a non-empty
// revision.date (the OpenAPI client models it as time.Time).
db2.featurerevisions.insertOne({
  organization: orgId,
  featureId: "e2e-flag",
  version: 1,
  baseVersion: 0,
  dateCreated: new Date(),
  dateUpdated: new Date(),
  datePublished: new Date(),
  publishedBy: {
    type: "api_key",
    id: "SECRET_API_KEY",
    name: "API",
    email: "",
  },
  comment: "seed",
  defaultValue: "false",
  rules: {},
  status: "published",
});


