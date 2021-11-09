import { assert } from "chai";
import { describe, it } from "mocha";
import { Credit } from "./Credit";

describe("Credit", () => {
  it("#fromApiObject", () => {
    const got = Credit.fromApiObject({
      BottomText: "contains gluten",
      MinorCategory: "farfalle",
      MajorCategory: "pasta",
      Name: "farfalle with sausage & asparagus",
      Order: "3",
      Project: "dinner",
    });
    assert.isNotEmpty(got);
    assert.equal(got.bottomText, "contains gluten");
    assert.equal(got.minorCategory, "farfalle");
    assert.equal(got.majorCategory, "pasta");
    assert.equal(got.name, "farfalle with sausage & asparagus");
    assert.equal(got.order, "3");
    assert.equal(got.project, "dinner");
  });

  it("#fromDatasetRow", () => {
    const got = Credit.fromDatasetRow(
      new Map([
        ["BottomText", "contains gluten"],
        ["MinorCategory", "farfalle"],
        ["MajorCategory", "pasta"],
        ["Name", "farfalle with sausage & asparagus"],
        ["Order", "3"],
        ["Project", "dinner"],
      ])
    );
    assert.isNotEmpty(got);
    assert.equal(got.bottomText, "contains gluten");
    assert.equal(got.minorCategory, "farfalle");
    assert.equal(got.majorCategory, "pasta");
    assert.equal(got.name, "farfalle with sausage & asparagus");
    assert.equal(got.order, "3");
    assert.equal(got.project, "dinner");
  });
});
