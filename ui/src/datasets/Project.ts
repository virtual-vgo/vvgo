export class Project {
    Name: string = "";
    Title: string = "";
    Season: string = "";
    Hidden: boolean = false;
    PartsReleased: boolean = false;
    PartsArchived: boolean = false;
    VideoReleased: boolean = false;
    Sources: string = "";
    Composers: string = "";
    Arrangers: string = "";
    Editors: string = "";
    Transcribers: string = "";
    Preparers: string = "";
    ClixBy: string = "";
    Reviewers: string = "";
    Lyricists: string = "";
    AdditionalContent: string = "";
    ReferenceTrack: string = "";
    ChoirPronunciationGuide: string = "";
    BannerLink: string = "";
    YoutubeLink: string = "";
    YoutubeEmbed: string = "";
    SubmissionDeadline: string = "";
    SubmissionLink: string = "";
    ReferenceTrackLink: string = "";

    static fromApiJSON(obj: object): Project {
        return obj as Project;
    }
}

export const latestProject = (projects: Project[] | undefined): Project | undefined =>
    projects?.filter(proj => proj.VideoReleased)
        .sort((a, b) => a.Name.localeCompare(b.Name))
        .pop();
