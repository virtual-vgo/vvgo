import { assert } from "chai";
import { describe, it } from "mocha";
import { Channel } from "./Channel";

describe("Channel", () => {
  it("#fromApiObject", () => {
    const got = Channel.fromApiObject({
      id: "cannelloni",
      name: "fettuccine",
      type: 42069,
    });
    assert.isNotEmpty(got);
    assert.equal(got.id, "cannelloni");
    assert.equal(got.name, "fettuccine");
    assert.equal(got.type, 42069);
  });
});
