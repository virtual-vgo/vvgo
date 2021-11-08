import { isEmpty, shuffle, uniqBy } from "lodash/fp";
import { useRef, useState } from "react";
import { Button, Col, FormControl, Row } from "react-bootstrap";
import ReactMarkdown from "react-markdown";
import { getSession } from "../../auth";
import { links } from "../../data/links";
import {
  MixtapeProject,
  Session,
  useGuildMemberLookup,
  useMixtapeProjects,
  useProjects,
  UserRole,
} from "../../datasets";
import { FancyProjectMenu, useMenuSelection } from "../shared/FancyProjectMenu";
import { CurrentMixtape } from "./NewProjectWorkflow";

const permaLink = (project: MixtapeProject) => `/mixtape/${project.Name}`;
const pathMatcher = /\/mixtape\/(.+)\/?/;

const searchProjects = (
  query: string,
  projects: MixtapeProject[]
): MixtapeProject[] => {
  return (projects ?? []).filter(
    (project) =>
      project.Name.toLowerCase().includes(query) ||
      project.channel?.toLowerCase().includes(query) ||
      project.hosts?.map((x) => x.toLowerCase()).includes(query)
  );
};

export const MemberDashboard = () => {
  const vvgoProjects = useProjects();
  const [mixtapeProjects, setMixtapeProjects] = useMixtapeProjects();
  const hosts = useGuildMemberLookup(
    (mixtapeProjects ?? []).flatMap((r) => r.hosts ?? [])
  );
  const [selected, setSelected] = useMenuSelection(
    mixtapeProjects ?? [],
    pathMatcher,
    permaLink,
    shuffle(mixtapeProjects).pop()
  );
  const me = getSession();

  const thisMixtape = vvgoProjects
    ?.filter((x) => x.Name == CurrentMixtape)
    .pop();

  const submissionDeadline =
    thisMixtape?.SubmissionDeadline ?? "the heat death of the universe";
  return (
    <div>
      <h3>
        <em>Hosts: final track submissions are due by{submissionDeadline}.</em>
      </h3>
      <Row className={"row-cols-1"}>
        <Col lg={3}>
          <FancyProjectMenu
            choices={mixtapeProjects ?? []}
            selected={selected}
            setSelected={setSelected}
            permaLink={permaLink}
            searchChoices={searchProjects}
            buttonContent={(proj) => (
              <div>
                {proj.title}
                <br />
                <small>
                  <em>{proj.resolveNicks(hosts).join(", ")}</em>
                </small>
              </div>
            )}
          />
        </Col>
        <Col lg={9}>
          {selected ? (
            <ProjectCard
              me={me}
              hostNicks={selected.resolveNicks(hosts)}
              project={selected}
              setProject={setSelected}
              allProjects={mixtapeProjects ?? []}
              setAllProjects={setMixtapeProjects}
            />
          ) : (
            <div />
          )}
        </Col>
      </Row>
    </div>
  );
};

const ProjectCard = (props: {
  me: Session;
  hostNicks: string[];
  project: MixtapeProject;
  setProject: (x: MixtapeProject) => void;
  allProjects: MixtapeProject[];
  setAllProjects: (x: MixtapeProject[]) => void;
}) => {
  const [showEdit, setShowEdit] = useState("");
  const blurbRef = useRef({} as HTMLTextAreaElement);

  let canEdit = false;
  switch (true) {
    case isEmpty(props.me.discordID):
      break;
    case props.me.roles?.includes(UserRole.ExecutiveDirector):
      canEdit = true;
      break;
    case props.project.hosts?.includes(props.me.discordID ?? ""):
      canEdit = true;
      break;
  }

  const onClickSubmit = () => {
    const proj = props.project;
    proj.blurb = blurbRef.current.value;
    setShowEdit("");
    proj.save().then((resp) => {
      props.setProject(proj);
      props.setAllProjects(
        uniqBy(
          (x) => x.Name,
          [...(resp.mixtapeProjects ?? []), ...props.allProjects]
        )
      );
    });
  };

  const blurb = props.project.blurb ?? "";
  return (
    <div>
      <h1>{props.project.title}</h1>
      <h4>
        Hosts: {props.hostNicks.join(", ")}
        <br />
        Channel: <em>{props.project.channel}</em>
      </h4>
      {showEdit == props.project.Name ? (
        <div className="mb-3">
          <FormControl
            ref={blurbRef}
            as={"textarea"}
            defaultValue={props.project.blurb}
            placeholder={"Description"}
          />
          <br />
          <a href={links.Help.Markdown}>Markdown Cheatsheet</a>
        </div>
      ) : (
        <ReactMarkdown>
          {blurb == "" ? "Project details coming soon!" : blurb}
        </ReactMarkdown>
      )}
      {canEdit ? (
        showEdit == props.project.Name ? (
          <Button
            type={"button"}
            variant={"outline-primary"}
            size={"sm"}
            onClick={onClickSubmit}
          >
            Submit
          </Button>
        ) : (
          <Button
            type={"button"}
            variant={"outline-primary"}
            size={"sm"}
            onClick={() => setShowEdit(props.project.Name)}
          >
            Edit
          </Button>
        )
      ) : (
        <div />
      )}
    </div>
  );
};

export default MemberDashboard;
