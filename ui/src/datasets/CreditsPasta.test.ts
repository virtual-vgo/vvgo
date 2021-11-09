import { assert } from "chai";
import { describe, it } from "mocha";
import { CreditsPasta } from "./CreditsPasta";

describe("CreditsPasta", () => {
  it("#fromApiObject", () => {
    const got = CreditsPasta.fromApiObject({
      WebsitePasta: "cannelloni",
      VideoPasta: "fettuccine",
      YoutubePasta: "rigatoni",
    });
    assert.isNotEmpty(got);
    assert.equal(got!.websitePasta, "cannelloni");
    assert.equal(got!.videoPasta, "fettuccine");
    assert.equal(got!.youtubePasta, "rigatoni");
  });
});
