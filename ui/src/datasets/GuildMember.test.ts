import { assert } from "chai";
import { describe, it } from "mocha";
import { DiscordUser, GuildMember } from "./GuildMember";

describe("GuildMember", () => {
  it("#fromApiObject", () => {
    const got = GuildMember.fromApiObject({
      user: { id: "cannelloni" },
      nick: "fettuccine",
      roles: ["rigatoni"],
    });
    assert.deepEqual(got.user, new DiscordUser("cannelloni"));
    assert.deepEqual(got.roles, ["rigatoni"]);
    assert.equal(got.nick, "fettuccine");
  });
});
