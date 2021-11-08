export class Part {
    Project: string = "";
    PartName: string = "";
    ScoreOrder: number = 0;
    SheetMusicFile: string = "";
    ClickTrackFile: string = "";
    ConductorVideo: string = "";
    PronunciationGuide: string = "";

    static fromApiJSON(obj: object): Part {
        return obj as Part;
    }
}
