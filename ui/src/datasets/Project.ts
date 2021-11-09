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
  Preparers = "";
  ClixBy = "";
  ReferenceTrack = "";
  BannerLink = "";
  YoutubeLink = "";
  YoutubeEmbed = "";
  SubmissionDeadline = "";
  SubmissionLink = "";
  BandcampAlbum = "";

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
