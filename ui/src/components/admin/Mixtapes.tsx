import { CSSProperties, useRef, useState } from "react";
import { Table } from "react-bootstrap";
import ReactMarkdown from "react-markdown";
import {
  GuildMember,
  MixtapeProject,
  useGuildMembers,
  useMixtapeProjects,
} from "../../datasets";

const ManageMixtapes = () => {
  return (
    <div>
      <h1 className="mb-3">Mixtapes</h1>
      <ProjectTable />
    </div>
  );
};

export default ManageMixtapes;

const ProjectTable = () => {
  const [projects] = useMixtapeProjects();
  const guildMembers = useGuildMembers();
  return (
    <div>
      <Table variant="dark" bordered size="small">
        {projects
          ?.sort((a, b) => a.id - b.id)
          .flatMap((proj, i) => [
            <thead key={`head-${i}`}>
              <tr>
                <td className="text-muted">id</td>
                <th>Title</th>
                <th>Name</th>
                <th>Mixtape</th>
                <th>Channel</th>
                <th>Hosts</th>
                <th />
              </tr>
            </thead>,
            <ProjectRow
              key={`body-${i}`}
              project={proj}
              guildMembers={guildMembers ?? []}
            />,
          ])}
      </Table>
    </div>
  );
};

const ProjectRow = (props: {
  project: MixtapeProject | undefined;
  guildMembers: GuildMember[];
}) => {
  const [project, setProject] = useState(props.project);
  const deleteProject = () => {
    project?.delete().then(() => setProject(undefined));
  };
  if (!project) return <div />;
  return (
    <tbody>
      <tr>
        <td className="text-muted">
          <em># {project.id}</em>
        </td>
        <EditField
          project={project}
          setProject={setProject}
          initValue={project.title}
          setField={(val, proj) => (proj.title = val)}
        />
        <EditField
          project={project}
          setProject={setProject}
          initValue={project.Name}
          setField={(val, proj) => (proj.Name = val)}
        />
        <EditField
          project={project}
          setProject={setProject}
          initValue={project.mixtape}
          setField={(val, proj) => (proj.mixtape = val)}
        />
        <EditField
          project={project}
          setProject={setProject}
          initValue={project.channel}
          setField={(val, proj) => (proj.channel = val)}
        />
        <td>
          {props.project?.resolveNicks(props.guildMembers ?? []).join(", ")}
        </td>
        <td>
          <span
            onClick={() => deleteProject()}
            style={{ cursor: "pointer" }}
            className="text-primary"
          >
            delete
          </span>
        </td>
      </tr>
      <tr>
        <td colSpan={7}>
          <ReactMarkdown>{props.project?.blurb ?? ""}</ReactMarkdown>
        </td>
      </tr>
    </tbody>
  );
};

const EditField = (props: {
  project: MixtapeProject;
  setProject: (proj: MixtapeProject) => void;
  initValue: string;
  setField: (val: string, proj: MixtapeProject) => void;
  as?: JSX.Element;
}) => {
  const [tdClassName, setTdClassName] = useState("");
  const inputStyle: CSSProperties = {
    height: "100%",
    width: "100%",
    textDecorationColor: "#000000",
    backgroundColor: "#ffffff",
  };
  const inputRef = useRef<HTMLInputElement>(null);

  const saveProject = () => {
    const newVal = inputRef.current?.value ?? props.initValue;
    if (props.project.id != 0 && props.initValue == newVal) return;

    const project = props.project;
    props.setField(newVal, project);
    (project.id == 0 ? project.create() : project.save()).then((resp) => {
      if (!resp.mixtapeProject) {
        setTdClassName("text-warning");
      } else {
        setTdClassName("");
        props.setProject(resp.mixtapeProject);
        console.log("saved project", resp.mixtapeProject);
      }
    });
  };

  return (
    <td className={tdClassName} onBlur={() => saveProject()}>
      <input
        style={inputStyle}
        ref={inputRef}
        type="text"
        defaultValue={props.initValue}
      />
    </td>
  );
};
