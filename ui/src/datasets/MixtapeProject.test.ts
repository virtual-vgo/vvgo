import { assert } from "chai";
import { describe, it } from "mocha";
import { MixtapeProject } from "./MixtapeProject";

describe("MixtapeProject", () => {
  it("#constructor", () => {
    const got = new MixtapeProject(
      "farfalle with sausage & asparagus",
      "pasta"
    );
    assert.equal(got.Name, "farfalle with sausage & asparagus");
    assert.equal(got.mixtape, "pasta");
  });

  it("#fromApiObject", () => {
    const got = MixtapeProject.fromApiObject({
      Name: "farfalle with sausage & asparagus",
      mixtape: "pasta",
      title: "farfalle-with-sausage-asparagus",
      blurb:
        "Farfalle, playfully referred to as bow tie pasta, soaks up just the right amount of sauce. Use it in Italian dinners, pasta salads and casseroles.",
      channel: "#farfalle-with-sausage-asparagus",
      hosts: ["only pasta"],
    });
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
    const proj = new MixtapeProject(
      "farfalle with sausage & asparagus",
      "pasta"
    );
    proj.title = "farfalle-with-sausage-asparagus";
    proj.blurb =
      "Farfalle, playfully referred to as bow tie pasta, soaks up just the right amount of sauce. Use it in Italian dinners, pasta salads and casseroles.";
    proj.channel = "#farfalle-with-sausage-asparagus";
    proj.hosts = ["only pasta"];
    assert.deepEqual(proj.toApiObject(), {
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
