import { assert } from "chai";
import { describe, it } from "mocha";
import { CreditsTable, CreditsTeamRow, CreditsTopic } from "./CreditsTable";

describe("CreditsTeamRow", () => {
  it("#fromApiObject", () => {
    const got = CreditsTeamRow.fromApiObject({
      Name: "Composers",
      Rows: [{ Name: "Comatose" }],
    });
    assert.isNotEmpty(got);
    assert.equal(got.Name, "Composers");
    assert.equal(got.Rows.length, 1);
    assert.equal(got.Rows[0].name, "Comatose");
  });
});

describe("CreditsTopicRow", () => {
  it("#fromApiObject", () => {
    const got = CreditsTopic.fromApiObject({
      Name: "Crew",
      Rows: [{ Name: "Composers" }],
    });
    assert.isNotEmpty(got);
    assert.equal(got.Name, "Crew");
    assert.equal(got.Rows.length, 1);
    assert.equal(got.Rows[0].Name, "Composers");
  });
});

describe("CreditsTable", () => {
  it("#fromApiArray", () => {
    const got = CreditsTable.fromApiArray([
      {
        Name: "Crew",
        Rows: [{ Name: "Composers" }],
      },
    ]);
    assert.isNotEmpty(got);
    assert.equal(got![0].Name, "Crew");
    assert.equal(got![0].Rows.length, 1);
    assert.equal(got![0].Rows[0].Name, "Composers");
  });
});
