import { isEmpty } from "lodash/fp";
import { lazy, Suspense } from "react";
import Col from "react-bootstrap/Col";
import Row from "react-bootstrap/Row";
import {
  latestProject,
  Project,
  useCreditsTable,
  useProjects,
} from "../datasets";
import { AlertUnreleasedProject } from "./shared/AlertUnreleasedProject";
import { FancyProjectMenu, useMenuSelection } from "./shared/FancyProjectMenu";
import { LoadingText } from "./shared/LoadingText";
import { ProjectHeader } from "./shared/ProjectHeader";
import { YoutubeIframe } from "./shared/YoutubeIframe";

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
            buttonContent={(proj: Project) => (
              <div>
                {proj.Title}
                {!proj.PartsReleased ? (
                  <em>
                    <small>
                      <br />
                      Unreleased
                    </small>
                  </em>
                ) : (
                  ""
                )}
                {!proj.VideoReleased ? (
                  <em>
                    <small>
                      <br />
                      In Production
                    </small>
                  </em>
                ) : (
                  ""
                )}
                {proj.VideoReleased ? (
                  <em>
                    <small>
                      <br />
                      Completed
                    </small>
                  </em>
                ) : (
                  ""
                )}
              </div>
            )}
          />
        </Col>
        <Col>
          {selected ? (
            <div className="mx-4">
              <AlertUnreleasedProject project={selected} />
              <ProjectHeader project={selected} />
              {selected.PartsArchived ? (
                selected.YoutubeEmbed ? (
                  <YoutubeIframe project={selected} />
                ) : (
                  <div className="text-center text-info">
                    <em>Video coming soon!</em>
                  </div>
                )
              ) : (
                <div />
              )}
              <ProjectCredits project={selected} />
            </div>
          ) : (
            <div />
          )}
        </Col>
      </Row>
    </div>
  );
};

const ProjectCredits = (props: { project: Project }) => {
  const Masonry = lazy(() => import("@mui/lab/Masonry"));
  const creditsTable = useCreditsTable(props.project);
  return (
    <div>
      {(creditsTable ?? []).map((topic) => (
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
              <Masonry
                columns={3}
                spacing={1}
                defaultHeight={450}
                defaultColumns={3}
                defaultSpacing={1}
              >
                {isEmpty(topic.Rows) ? (
                  <div />
                ) : (
                  topic.Rows.map((team) => (
                    <Col key={team.Name} lg={4}>
                      <h5>{team.Name}</h5>
                      <ul className="list-unstyled">
                        {team.Rows.map((credit, i) => (
                          <li key={i}>
                            {credit.name} <small>{credit.bottomText}</small>
                          </li>
                        ))}
                      </ul>
                    </Col>
                  ))
                )}
              </Masonry>
            </Suspense>
          </Row>
        </Row>
      ))}
    </div>
  );
};

export default Projects;
