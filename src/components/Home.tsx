import Col from "react-bootstrap/Col";
import Row from "react-bootstrap/Row";
import {
  Highlight,
  latestProject,
  Project,
  useHighlights,
  useProjects,
} from "../datasets";
import { randElement } from "../utils";
import { ProjectBanner } from "./shared/ProjectBanner";
import { YoutubeIframe } from "./shared/YoutubeIframe";

export const Home = () => {
  const highlights = useHighlights();
  const projects = useProjects();
  const highlight = highlights ? randElement(highlights) : undefined;

  const latest = projects ? latestProject(projects) : undefined;
  return (
    <div>
      <Row>
        <Col lg={7} md={12}>
          <ProjectBanner project={latest} />
          <YoutubeIframe project={latest} />
        </Col>
        <Col>
          <div className={"col mt-2"}>
            <h3>Latest Releases</h3>
            <LatestReleases projects={projects} />
            <h3>Member Highlights</h3>
            <MemberHighlight highlight={highlight} />
          </div>
          <div className="row justify-content-md-center text-center m-2">
            <div className="col text-center mt-2">
              <p>
                If you would like to join our orchestra or get more information
                about our current projects, please join us on{" "}
                <a href="https://discord.gg/mt8EZdsrc6">Discord!</a>
              </p>
            </div>
          </div>
        </Col>
      </Row>
    </div>
  );
};

const LatestReleases = (props: { projects: Project[] | undefined }) => {
  if (!props.projects) return <div />;
  const projects = props.projects.filter((p) => p.VideoReleased).slice(0, 3);

  const ProjectRow = (props: { project: Project }) => {
    return (
      <tr>
        <td>
          <a href={props.project.YoutubeLink} className="text-light">
            {props.project.Title} <br /> {props.project.Sources}
          </a>
        </td>
      </tr>
    );
  };

  return (
    <table className="table text-light clickable">
      <tbody>
        {projects.map((project) => (
          <ProjectRow key={project.Name} project={project} />
        ))}
      </tbody>
    </table>
  );
};

const MemberHighlight = (props: { highlight?: Highlight }) => {
  if (!props.highlight) return <div />;
  const { source, alt } = props.highlight;
  return (
    <table className="table text-light">
      <tbody>
        <tr>
          <td>
            <img src={source} width="100%" alt={alt} />
          </td>
        </tr>
      </tbody>
    </table>
  );
};

export default Home;
