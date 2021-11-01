import _ = require("lodash");
import React = require("react");
import Button from "react-bootstrap/Button";
import ButtonGroup from "react-bootstrap/ButtonGroup";
import FormControl from "react-bootstrap/FormControl";
import {ShowHideToggle} from "./ShowHideToggle";

export function useMenuSelection<T extends { Name: string }>(
    choices: T[],
    pathMatcher: RegExp,
    permaLink: (x: T) => string,
    defaultChoice: T,
): [T, React.Dispatch<React.SetStateAction<T>>] {
    const [selected, setSelected] = React.useState(null as T);

    window.onpopstate = (event) => {
        if (event.state) setSelected(event.state);
    };

    if (selected) return [selected, setSelected];
    if (_.isEmpty(choices)) return [selected, setSelected];

    let want = defaultChoice;

    const params = new URLSearchParams(document.location.search);
    if (!_.isEmpty(params.get("name")))
        want = choices.filter(p => p.Name == params.get("name")).pop();

    const pathMatch = document.location.pathname.match(pathMatcher);
    if (pathMatch && pathMatch.length == 2 && pathMatch[1].length > 0)
        want = choices.filter(p => p.Name == pathMatch[1]).pop();

    if (want) {
        setSelected(want);
        window.history.pushState(want, "", permaLink(want));
    }

    return [selected, setSelected];
}

export function FancyProjectMenu<T extends { Name: string }>(props: {
    choices: T[],
    permaLink: (x: T) => string,
    selected: T,
    setSelected: React.Dispatch<React.SetStateAction<T>>,
    buttonContent: (x: T) => JSX.Element
    searchChoices?: (q: string, choices: T[]) => T[],
    toggles?: Array<{ hidden: boolean, title: string, filter: (on: boolean, x: T) => boolean }>,
}) {
    const [searchInput, setSearchInput] = React.useState("");
    const menuToggles = useToggles(props.toggles);

    const searcher = _.defaultTo(props.searchChoices, () => props.choices);
    const filter = (x: T) => _.isEmpty(menuToggles.filter(t => t.state)) ||
        menuToggles.filter(t => t.state)
            .map(t => t.filter(t.state, x))
            .reduce((a, b) => (a && b));
    const wantChoices = searcher(searchInput, props.choices).filter(filter);

    const onClickChoice = (want: T) => {
        const params = new URLSearchParams({name: want.Name});
        window.history.pushState(params, "", props.permaLink(want));
        props.setSelected(want);
    };

    return <div>
        <MenuToggles toggles={menuToggles}/>
        <SearchBox hidden={props.searchChoices == null} setSearchInput={setSearchInput}/>
        <div className="d-grid">
            <ButtonGroup vertical className="m-2">
                {wantChoices.map(want =>
                    <Button
                        variant={props.selected && props.selected.Name == want.Name ? "light" : "outline-light"}
                        key={want.Name}
                        onClick={() => onClickChoice(want)}>
                        {props.buttonContent(want)}
                    </Button>)}
            </ButtonGroup>
        </div>
    </div>;
}

export type Toggle<T> = {
    hidden: boolean,
    title: string,
    state: boolean,
    setState: (b: boolean) => void,
    filter: (on: boolean, x: T) => boolean
}

export function useToggles<T>(toggles: Array<{
    hidden: boolean,
    title: string,
    filter: (on: boolean, x: T) => boolean
}>): Toggle<T>[] {
    const [toggleState, setToggleState] = React.useState(0);
    return _.defaultTo(toggles, []).map((t, i) => ({
        ...t,
        state: (toggleState & (1 << i)) == (1 << i),
        setState: (x: boolean) => setToggleState(x ? toggleState | (1 << i) : toggleState),
    }));
}

export function MenuToggles<T>(props: { toggles: Toggle<T>[] }): JSX.Element {
    return _.isEmpty(props.toggles) ?
        <div/> :
        <div className={"d-flex flex-row justify-content-center"}>
            {props.toggles.map((toggle) =>
                <ShowHideToggle
                    title={toggle.title}
                    state={toggle.state}
                    setState={toggle.setState}/>)}
        </div>;
}

export function SearchBox(props: { hidden: boolean, setSearchInput: (q: string) => void }): JSX.Element {
    return props.hidden ?
        <div/> :
        <div className="d-flex flex-row justify-content-center">
            <FormControl
                className="m-2"
                placeholder="search projects"
                onChange={(event) => props.setSearchInput(event.target.value.toLowerCase())}/>
        </div>;
}
