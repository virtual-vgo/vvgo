import { assert } from "chai";
import { describe, it } from "mocha";
import { CreditsPasta } from "./CreditsPasta";

describe("CreditsPasta", () => {
  it("#fromApiObject", () => {
    const got = CreditsPasta.fromApiObject({
      Name: "farfalle with sausage & asparagus",
      mixtape: "fettuccine",
      title: "rigatoni",
      blurb: "rigatoni",
      channel: "rigatoni",
      hosts: "rigatoni",
    });
    assert.isNotEmpty(got);
    assert.equal(got!.websitePasta, "cannelloni");
    assert.equal(got!.videoPasta, "fettuccine");
    assert.equal(got!.youtubePasta, "rigatoni");
  });
});
