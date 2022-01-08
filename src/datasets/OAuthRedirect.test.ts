import { assert } from "chai";
import { describe, it } from "mocha";
import { OAuthRedirect } from "./OAuthRedirect";

describe("OAuthRedirect", () => {
  it("#fromApiObject", () => {
    const got = OAuthRedirect.fromApiObject({
      DiscordURL: "no u",
      State: "of despair",
      Secret: "sauce",
    });
    assert.isNotEmpty(got);
    assert.deepEqual(got!.DiscordURL, "no u");
    assert.deepEqual(got!.State, "of despair");
    assert.equal(got!.Secret, "sauce");
  });
});
