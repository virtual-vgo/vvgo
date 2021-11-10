import { assert } from "chai";
import { describe, it } from "mocha";
import { DiscordUser, GuildMember } from "./GuildMember";

describe("GuildMember", () => {
  it("#fromApiObject", () => {
    const got = GuildMember.fromApiObject({
      user: { id: "cannelloni", username: "linguini" },
      nick: "fettuccine",
      roles: ["rigatoni"],
    });
    assert.deepEqual(got.user.id, "cannelloni");
    assert.equal(got.user.username, "linguini");
    assert.deepEqual(got.roles, ["rigatoni"]);
    assert.equal(got.nick, "fettuccine");
  });
});
