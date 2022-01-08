import { assert } from "chai";
import { describe, it } from "mocha";
import { Director } from "./Director";

describe("Director", () => {
  it("#fromApiObject", () => {
    const got = Director.fromDatasetRow(
      new Map([
        ["Name", "farfalle with sausage & asparagus"],
        [
          "Epithet",
          "This pasta dish comes together fast on hectic nights and makes wonderful leftovers",
        ],
        ["Affiliations", "20 Fantastic Farfalle Recipes"],
        [
          "Blurb",
          "Farfalle, playfully referred to as bow tie pasta, soaks up just the right amount of sauce. Use it in Italian dinners, pasta salads and casseroles.",
        ],
        ["Icon", "/farfalle.png"],
      ])
    );
    assert.isNotEmpty(got);
    assert.equal(got.name, "farfalle with sausage & asparagus");
    assert.equal(
      got.epithet,
      "This pasta dish comes together fast on hectic nights and makes wonderful leftovers"
    );
    assert.equal(got.affiliations, "20 Fantastic Farfalle Recipes");
    assert.equal(
      got.blurb,
      "Farfalle, playfully referred to as bow tie pasta, soaks up just the right amount of sauce. Use it in Italian dinners, pasta salads and casseroles."
    );
    assert.equal(got.icon, "/farfalle.png");
  });
});
