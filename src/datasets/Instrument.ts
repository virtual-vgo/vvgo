import { DatasetRow } from "./Dataset";

export class Instrument {
  partName = "";
  partNameStripped = "";
  partID = "";
  instrumentIndex = "";

  static fromDatasetRow(data: DatasetRow): Instrument {
    const instrument = new Instrument();
    instrument.partName = data.get("CreditedRole/PartName") ?? "";
    instrument.partNameStripped = instrument.partName
      .replace(/ [0-9]/g, "")
      .replace(/2ND /g, "")
      .replace(/3RD /g, "")
      .trim();
    instrument.partID = data.get("CombinedPart#") ?? "";
    instrument.instrumentIndex = data.get("Instr.Index") ?? "";
    return instrument;
  }
}
