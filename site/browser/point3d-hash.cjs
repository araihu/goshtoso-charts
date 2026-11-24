const assert = require("node:assert/strict");
const { createHash } = require("node:crypto");
const hashes = require("../internal/pages/testdata/point3d-hashes.json");

function checkPoint3DHash(name, points) {
  const triples = points.map((point) => point.map((coordinate) => {
    assert.ok(Number.isFinite(coordinate), "chart coordinates must be finite numbers");
    // Match Go's math.Round for negative ties as well as positive ones.
    const scaled = Math.sign(coordinate) * Math.round(Math.abs(coordinate) * 1e9);
    assert.ok(Number.isSafeInteger(scaled));
    return scaled;
  }));
  const got = createHash("sha256").update(JSON.stringify(triples)).digest("hex");
  assert.equal(got, hashes[name], name);
}

module.exports = { checkPoint3DHash };
