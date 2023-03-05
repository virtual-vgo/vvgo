import { Button, Col, FormControl, Row } from "react-bootstrap";
import { FancyProjectMenu, useMenuSelection } from "../shared/FancyProjectMenu";
import { GuildMember, MixtapeProject, Session, UserRole, useGuildMembers, useMixtapeProjects } from "../../datasets";
import { isEmpty, uniqBy } from "lodash/fp";
import { useRef, useState } from "react";

import { LinkUser } from "../shared/LinkChannel";
import ReactMarkdown from "react-markdown";
import { getSession } from "../../auth";
import { links } from "../../data/links";

const permaLink = (project: MixtapeProject) => `/mixtape/${project.Name}`;
const pathMatcher = /\/mixtape\/(.+)\/?/;

const searchProjects = (query: string, projects: MixtapeProject[]): MixtapeProject[] => {
    return (projects ?? []).filter(
        (project) =>
            project.Name.toLowerCase().includes(query) ||
            project.channel?.toLowerCase().includes(query) ||
            project.hosts?.map((x) => x.toLowerCase()).includes(query)
    );
};

export const MemberDashboard = () => {
    const [mixtapeProjects, setMixtapeProjects] = useMixtapeProjects();
    const guildMembers = useGuildMembers() ?? [];

    const filteredProjects =
        mixtapeProjects?.filter((p) => p.title !== "").sort((a, b) => (a.title > b.title ? 1 : -1)) ?? [];
    const { selected, onSelect } = useMenuSelection(filteredProjects, pathMatcher, permaLink, filteredProjects[0]);
    const me = getSession();

    return (
        <div>
            <Row className={"row-cols-1"}>
                <Col lg={3}>
                    <FancyProjectMenu
                        choices={filteredProjects}
                        selected={selected}
                        onSelect={onSelect}
                        permaLink={permaLink}
                        searchChoices={searchProjects}
                        buttonContent={(proj) => (
                            <div>
                                {proj.title}
                                <br />
                                <small>
                                    <em>{proj.resolveNicks(guildMembers).join(", ")}</em>
                                </small>
                            </div>
                        )}
                    />
                </Col>
                <Col lg={9}>
                    <ProjectCard
                        me={me}
                        guildMembers={guildMembers}
                        project={selected}
                        onSelect={onSelect}
                        allProjects={mixtapeProjects ?? []}
                        setAllProjects={setMixtapeProjects}
                    />
                </Col>
            </Row>
        </div>
    );
};

const ProjectCard = (props: {
    me: Session;
    guildMembers: GuildMember[];
    project: MixtapeProject | undefined;
    onSelect: (x: MixtapeProject) => void;
    allProjects: MixtapeProject[];
    setAllProjects: (x: MixtapeProject[]) => void;
}) => {
    const [showEdit, setShowEdit] = useState("");
    const blurbRef = useRef({} as HTMLTextAreaElement);
    if (!props.project) return <div />;
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
        if (!props.project) return;
        const proj = props.project;
        proj.blurb = blurbRef.current.value;
        setShowEdit("");
        proj.save().then((resp) => {
            props.onSelect(proj);
            const allProjects = uniqBy((x) => x.Name, [...(resp.mixtapeProjects ?? []), ...props.allProjects]);
            props.setAllProjects(allProjects);
        });
    };

    const hosts = isEmpty(props.project.hosts) ? (
        <div />
    ) : (
        <div>
            Hosts:{" "}
            <ul>
                {props.guildMembers
                    .filter((m) => props.project?.hosts?.includes(m.user.id))
                    .filter((m) => m.user && m.user.username && m.user.username != "")
                    .map((m, i) => (
                        <li key={i}>
                            <LinkUser member={m} />{" "}
                        </li>
                    ))}
            </ul>
        </div>
    );

    const blurbInput = (
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
    );

    const blurbContent = (
        <ReactMarkdown>
            {props.project.blurb == "" ? "Project details coming soon!" : props.project.blurb}
        </ReactMarkdown>
    );

    let bottomButtons: JSX.Element[] = [];
    if (showEdit == props.project.Name)
        bottomButtons = [
            <Button key={1} type={"button"} variant={"outline-primary"} size={"sm"} onClick={onClickSubmit}>
                Submit
            </Button>,
            <Button key={2} type={"button"} variant={"outline-primary"} size={"sm"} onClick={() => setShowEdit("")}>
                Cancel
            </Button>,
        ];
    else if (canEdit)
        bottomButtons = [
            <Button
                key={1}
                type={"button"}
                variant={"outline-primary"}
                size={"sm"}
                onClick={() => props.project && setShowEdit(props.project.Name)}
            >
                Edit
            </Button>,
        ];

    return (
        <div>
            <h1>{props.project.title}</h1>
            <h4>
                {hosts}
                <br />
                Channel: <em>{props.project.channel}</em>
            </h4>
            {showEdit == props.project.Name ? blurbInput : blurbContent}
            {bottomButtons}
        </div>
    );
};

export default MemberDashboard;
