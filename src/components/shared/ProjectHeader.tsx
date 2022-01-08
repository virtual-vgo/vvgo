import { isEmpty } from "lodash/fp";
import Col from "react-bootstrap/Col";
import Row from "react-bootstrap/Row";
import { Project } from "../../datasets";
import { ProjectBanner } from "./ProjectBanner";

export const ProjectHeader = (props: { project: Project }) => (
  <Row className="row-cols-1">
    <Col className="text-center">
      <ProjectBanner project={props.project} />
      {props.project.Composers}
      <br />
      <small>{props.project.Arrangers}</small>
      {props.project.PartsArchived ||
      isEmpty(props.project.SubmissionDeadline) ? (
        <div />
      ) : (
        <div className="m-2">
          <h4>
            <strong>Submission Deadline:</strong>{" "}
            <em>{props.project.SubmissionDeadline} (Hawaii Time)</em>
          </h4>
        </div>
      )}
    </Col>
  </Row>
);
