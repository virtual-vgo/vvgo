export class Project {
  Name = "";
  Title = "";
  Season = "";
  Hidden = false;
  PartsReleased = false;
  PartsArchived = false;
  VideoReleased = false;
  Sources = "";
  Composers = "";
  Arrangers = "";
  Editors = "";
  Transcribers = "";
  Preparers = "";
  ClixBy = "";
  Reviewers = "";
  Lyricists = "";
  AdditionalContent = "";
  ReferenceTrack = "";
  ChoirPronunciationGuide = "";
  BannerLink = "";
  YoutubeLink = "";
  YoutubeEmbed = "";
  SubmissionDeadline = "";
  SubmissionLink = "";
  ReferenceTrackLink = "";

  static fromApiObject(obj: object): Project {
    return obj as Project;
  }
}

export const latestProject = (
  projects: Project[] | undefined
): Project | undefined =>
  projects
    ?.filter((proj) => proj.VideoReleased)
    .sort((a, b) => a.Name.localeCompare(b.Name))
    .pop();
