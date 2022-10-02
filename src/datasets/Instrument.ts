import { DatasetRow } from "./Dataset";

export class Instrument {
  partName = "";
  partNameStripped = "";
  partID = "";

  static fromDatasetRow(data: DatasetRow): Instrument {
    const instrument = new Instrument();
    instrument.partName = data.get("Credited Role/Part Name") ?? "";
    instrument.partNameStripped = instrument.partName
      .replace(/[0-9]/g, "")
      .trim();
    instrument.partID = data.get("Combined Part #") ?? "";
    return instrument;
  }
}
