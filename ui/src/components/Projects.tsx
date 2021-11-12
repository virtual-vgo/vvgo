import { isEmpty } from "lodash/fp";
import { lazy, Suspense } from "react";
import Col from "react-bootstrap/Col";
import Row from "react-bootstrap/Row";
import { GuildChannel } from "../static/discord";
import {
  CreditsTeamRow,
  CreditsTopic,
  latestProject,
  Project,
  useCreditsTable,
  useProjects,
} from "../datasets";
import { AlertUnreleasedProject } from "./shared/AlertUnreleasedProject";
import { FancyProjectMenu, useMenuSelection } from "./shared/FancyProjectMenu";
import { LinkChannel } from "./shared/LinkChannel";
import { LoadingText } from "./shared/LoadingText";
import { ProjectHeader } from "./shared/ProjectHeader";
import { YoutubeIframe } from "./shared/YoutubeIframe";

const Masonry = lazy(() => import("@mui/lab/Masonry"));

const permaLink = (project: Project) => `/projects/${project.Name}`;
const pathMatcher = /\/projects\/(.+)\/?/;

const searchProjects = (
  query: string,
  projects: Project[] | undefined
): Project[] => {
  return (projects ?? []).filter(
    (r) =>
      r.Name.toLowerCase().includes(query) ||
      r.Title.toLowerCase().includes(query) ||
      r.Sources.toLowerCase().includes(query)
  );
};

export const Projects = () => {
  const allProjects = useProjects();
  const allowedProjects = (allProjects ?? []).filter((r) => !r.Hidden);
  const [selected, setSelected] = useMenuSelection(
    allowedProjects,
    pathMatcher,
    permaLink,
    latestProject(allowedProjects)
  );

  const buttonContent = (proj: Project) => {
    return (
      <div>
        {proj.Title}
        <em>
          <small>
            {!proj.PartsReleased ? <div>Unreleased</div> : <div />}
            {!proj.VideoReleased ? <div>In Production</div> : <div />}
            {proj.VideoReleased ? <div>Completed</div> : <div />}
          </small>
        </em>
      </div>
    );
  };

  if (!allProjects) return <LoadingText />;
  return (
    <div>
      <Row>
        <Col lg={3}>
          <FancyProjectMenu
            choices={allowedProjects}
            selected={selected}
            setSelected={setSelected}
            permaLink={permaLink}
            searchChoices={searchProjects}
            buttonContent={buttonContent}
          />
        </Col>
        <Col>
          <ProjectPage project={selected} />
        </Col>
      </Row>
    </div>
  );
};

const ProjectPage = (props: { project: Project | undefined }) => {
  if (props.project == undefined) return <LoadingText />;
  return (
    <div className="mx-4">
      <AlertUnreleasedProject project={props.project} />
      <AlertInProduction project={props.project} />
      <ProjectHeader project={props.project} />
      <YoutubeIframe project={props.project} />
      <ProjectCredits project={props.project} />
    </div>
  );
};

const AlertInProduction = (props: { project: Project | undefined }) => {
  if (props.project == undefined) return <div />;
  if (props.project.VideoReleased) return <div />;
  if (!props.project.PartsArchived) return <div />;
  return (
    <div className="text-muted mb-4 fa-border">
      <h2 className="m-2">
        <em>Hey beautiful!</em> 😉{" "}
        <em>
          This project is still in production, but we are no longer accepting
          submissions. Stay tuned for upcoming release news in{" "}
          <LinkChannel channel={GuildChannel.Announcements} />.
        </em>{" "}
        😘
      </h2>
    </div>
  );
};

const ProjectCredits = (props: { project: Project }) => {
  const creditsTable = useCreditsTable(props.project);
  return (
    <div>
      {creditsTable?.map((topic) => (
        <Row key={topic.Name}>
          <Row>
            <Col className="text-center">
              <h2>
                <strong>— {topic.Name} —</strong>
              </h2>
            </Col>
          </Row>
          <Row>
            <Suspense fallback={<LoadingText />}>
              <CreditsTopicMasonry topic={topic} />
            </Suspense>
          </Row>
        </Row>
      ))}
    </div>
  );
};

const CreditsTopicMasonry = (props: { topic: CreditsTopic }) => {
  if (isEmpty(props.topic.Rows)) return <div />;

  return (
    <Masonry defaultHeight={450} columns={{ md: 3, sm: 1 }} spacing={1}>
      {props.topic.Rows.map((team, i) => (
        <TeamCredits key={i} team={team} />
      ))}
    </Masonry>
  );
};

const TeamCredits = (props: { team: CreditsTeamRow }) => {
  return (
    <div className="text-center">
      <h5>{props.team.Name}</h5>
      <ul className="list-unstyled">
        {props.team.Rows.map((credit, i) => (
          <li key={i}>
            {credit.name} <small>{credit.bottomText}</small>
          </li>
        ))}
      </ul>
    </div>
  );
};

export default Projects;
