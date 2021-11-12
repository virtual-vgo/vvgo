import { assert } from "chai";
import { describe, it } from "mocha";
import { Resources } from "./Resources";

describe("Resources", () => {
  const db = new Resources("testing", "http://localhost:42069");

  describe("#channels", () => {
    it("#list", () => {
      db.channels.list().then((channels) => {
        assert.isNotEmpty(channels);
        assert.equal(channels![0].name, "cannelloni");
        assert.equal(channels![0].type, "fettuccine");
        assert.equal(channels![0].id, "rigatoni");
      });
    });
  });
});
