import { isEmpty, uniq } from "lodash/fp";
import { CSSProperties, useRef, useState } from "react";
import { Dropdown, Table, Toast } from "react-bootstrap";
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
  const [projects, setProjects] = useMixtapeProjects();
  const guildMembers = useGuildMembers();
  return (
    <div>
      <Table variant="dark" bordered size="small">
        <CreateProjectRow projects={projects} setProjects={setProjects} />
        {projects
          ?.sort((a, b) => b.id - a.id)
          .map((proj) => (
            <>
              <thead key={`${proj.id}-head`}>
                <tr>
                  <td className="text-muted">id</td>
                  <th>Title</th>
                  <th>Permalink</th>
                  <th>Mixtape</th>
                  <th>Channel</th>
                  <th>Hosts</th>
                  <th />
                </tr>
              </thead>
              <ProjectRow
                key={`${proj.id}-body`}
                projects={projects}
                setProjects={setProjects}
                thisProject={proj}
                guildMembers={guildMembers ?? []}
              />
            </>
          ))}
      </Table>
    </div>
  );
};

const CreateProjectRow = (props: {
  projects: MixtapeProject[] | undefined;
  setProjects: (projs: MixtapeProject[]) => void;
}) => {
  const createProject = () => {
    new MixtapeProject().create().then((resp) => {
      if (resp.mixtapeProject)
        props.setProjects([resp.mixtapeProject, ...(props.projects ?? [])]);
    });
  };
  return (
    <>
      <thead>
        <tr>
          <td className="text-muted">id</td>
          <th>Title</th>
          <th>Permalink</th>
          <th>Mixtape</th>
          <th>Channel</th>
          <th>Hosts</th>
          <th />
        </tr>
      </thead>
      <tbody>
        <tr>
          <td
            colSpan={7}
            className="text-underline text-primary"
            style={{ cursor: "pointer" }}
            onClick={createProject}
          >
            + new project
          </td>
        </tr>
      </tbody>
    </>
  );
};

const ProjectRow = (props: {
  thisProject: MixtapeProject | undefined;
  projects: MixtapeProject[] | undefined;
  setProjects: (projs: MixtapeProject[]) => void;
  guildMembers: GuildMember[];
}) => {
  const [project, setProject] = useState(props.thisProject);
  const deleteProject = () => {
    project?.delete().then(() => {
      setProject(undefined);
      props.setProjects(
        props.projects?.filter((p) => p.id != project.id) ?? []
      );
    });
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
        <EditHosts
          project={project}
          setProject={setProject}
          guildMembers={props.guildMembers}
        />
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
          <ReactMarkdown>{props.thisProject?.blurb ?? ""}</ReactMarkdown>
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
}) => {
  const [tdClassName, setTdClassName] = useState("");
  const inputStyle: CSSProperties = {
    height: "100%",
    width: "100%",
  };
  const inputRef = useRef<HTMLInputElement>(null);

  const saveProject = () => {
    const newVal = inputRef.current?.value ?? props.initValue;
    if (props.project.id != 0 && props.initValue == newVal) return;

    const project = props.project;
    props.setField(newVal, project);
    (project.id == 0 ? project.create() : project.save())
      .then((resp) => {
        if (!resp.mixtapeProject) {
          setTdClassName("text-warning");
        } else {
          setTdClassName("");
          props.setProject(resp.mixtapeProject);
          console.log("saved project", resp.mixtapeProject);
        }
      })
      .catch(() => setTdClassName("text-warning"));
  };

  return (
    <td
      className={tdClassName}
      onKeyUp={(e) => {
        if (["Enter", "Escape"].includes(e.code)) saveProject();
      }}
      onBlur={() => saveProject()}
    >
      <input
        style={inputStyle}
        ref={inputRef}
        type="text"
        defaultValue={props.initValue}
      />
    </td>
  );
};

const EditHosts = (props: {
  project: MixtapeProject;
  setProject: (proj: MixtapeProject | undefined) => void;
  guildMembers: GuildMember[];
}) => {
  const [showToast, setShowToast] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");

  const addHost = (member: GuildMember | undefined) => {
    if (!member) return;
    const project = props.project;
    project.hosts = uniq([...project.hosts, member.user.id]);
    project.save().then((resp) => {
      if (resp.mixtapeProject) props.setProject(resp.mixtapeProject);
    });
  };

  const rmHost = (member: GuildMember | undefined) => {
    if (!member) return;
    const project = props.project;
    project.hosts = project.hosts.filter((id) => id != member.user.id);
    project.save().then((resp) => {
      if (resp.mixtapeProject) props.setProject(resp.mixtapeProject);
    });
  };

  const projectMembers = props.guildMembers.filter((m) =>
    props.project.hosts.includes(m.user.id)
  );

  const filteredMembers =
    isEmpty(props.guildMembers) || isEmpty(searchQuery) || !showToast
      ? []
      : props.guildMembers
          ?.filter((m) => !props.project.hosts.includes(m.user.id))
          .filter(
            (m) =>
              m.user.username.toLowerCase().includes(searchQuery) ||
              m.nick.toLowerCase().includes(searchQuery) ||
              m.user.id.toString().includes(searchQuery)
          );

  const hostList = projectMembers.map((m, i) => (
    <li key={i}>
      {m.displayName()}{" "}
      <span
        className="text-primary text-decoration-underline"
        style={{ cursor: "pointer" }}
        onClick={() => rmHost(m)}
      >
        remove
      </span>
    </li>
  ));

  return (
    <td>
      <ul>
        {isEmpty(hostList) ? <li>This project has no hosts.</li> : hostList}
      </ul>
      <input
        placeholder="search nicks"
        defaultValue={searchQuery}
        onChange={(e) => {
          setShowToast(true);
          setSearchQuery(e.target.value?.toLowerCase() ?? "");
        }}
        onKeyUp={(e) => ["Escape"].includes(e.code) && setShowToast(false)}
      />
      <div>
        <Toast>
          {filteredMembers
            ?.filter((m) => !isEmpty(m.nick))
            .filter((m) => !props.project.hosts.includes(m.user.id))
            .slice(0, 5)
            .map((m, i) => (
              <Dropdown.Item key={i} onClick={() => addHost(m)}>
                {m.nick}
              </Dropdown.Item>
            ))}
        </Toast>
      </div>
    </td>
  );
};
