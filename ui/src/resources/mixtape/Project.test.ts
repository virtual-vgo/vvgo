import { assert } from "chai";
import { describe, it } from "mocha";
import { Project } from "./Project";

describe("Project", () => {
  it("#fromApiObject", () => {
    const got = Project.fromApiObject({
      id: 42069,
      Name: "farfalle with sausage & asparagus",
      mixtape: "pasta",
      title: "farfalle-with-sausage-asparagus",
      blurb:
        "Farfalle, playfully referred to as bow tie pasta, soaks up just the right amount of sauce. Use it in Italian dinners, pasta salads and casseroles.",
      channel: "#farfalle-with-sausage-asparagus",
      hosts: ["only pasta"],
    });
    assert.equal(got.id, 42069);
    assert.equal(got.Name, "farfalle with sausage & asparagus");
    assert.equal(got.mixtape, "pasta");
    assert.equal(got.title, "farfalle-with-sausage-asparagus");
    assert.equal(
      got.blurb,
      "Farfalle, playfully referred to as bow tie pasta, soaks up just the right amount of sauce. Use it in Italian dinners, pasta salads and casseroles."
    );
    assert.equal(got.channel, "#farfalle-with-sausage-asparagus");
    assert.deepEqual(got.hosts, ["only pasta"]);
  });

  it("#toApiObject", () => {
    const proj = new Project();
    proj.Name = "farfalle with sausage & asparagus";
    proj.mixtape = "pasta";
    proj.title = "farfalle-with-sausage-asparagus";
    proj.blurb =
      "Farfalle, playfully referred to as bow tie pasta, soaks up just the right amount of sauce. Use it in Italian dinners, pasta salads and casseroles.";
    proj.channel = "#farfalle-with-sausage-asparagus";
    proj.hosts = ["only pasta"];
    assert.deepEqual(proj.saveArgs(), {
      Name: "farfalle with sausage & asparagus",
      mixtape: "pasta",
      title: "farfalle-with-sausage-asparagus",
      blurb:
        "Farfalle, playfully referred to as bow tie pasta, soaks up just the right amount of sauce. Use it in Italian dinners, pasta salads and casseroles.",
      channel: "#farfalle-with-sausage-asparagus",
      hosts: ["only pasta"],
    });
  });
});
