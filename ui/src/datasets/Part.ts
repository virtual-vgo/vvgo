export class Part {
    Project = "";
    PartName = "";
    ScoreOrder = 0;
    SheetMusicFile = "";
    ClickTrackFile = "";
    ConductorVideo = "";
    PronunciationGuide = "";

    static fromApiObject(obj: object): Part {
        return obj as Part;
    }
}
