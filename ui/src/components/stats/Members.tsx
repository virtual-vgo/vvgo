import {flatRollup} from "d3";
import {isEmpty} from "lodash/fp";
import {VegaLite} from "react-vega";
import {useCredits} from "../../datasets";

export const Members = () => {
    const credits = (useCredits() ?? []).filter(x => !isEmpty(x.name));
    if (isEmpty(credits)) return <div/>;

    const topCredited = flatRollup(credits, g => g.length, p => p.name)
        .map(([name, count]) => ({name, count}))
        .sort((a, b) => a.count - b.count)
        .reverse()
        .slice(0, 10);

    const performers = credits.filter(x => x.majorCategory != "CREW");
    const topCreditedPerformers = flatRollup(performers, g => g.length, p => p.name)
        .map(([name, count]) => ({name, count}))
        .sort((a, b) => a.count - b.count)
        .reverse()
        .slice(0, 10);

    const topUniqueParts = flatRollup(performers, g => new Set(g).size, p => p.name)
        .map(([name, count]) => ({name, count}))
        .sort((a, b) => a.count - b.count)
        .reverse()
        .slice(0, 10);

    return <div>
        <div className={"d-grid justify-content-center"}>
            <h1>Member Stats</h1>
            <VegaLite
                data={{data: topCredited}}
                spec={{
                    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
                    title: "Top Credited Overall",
                    background: "transparent",
                    data: {name: "data"},
                    config: {title: {color: "white", fontSize: 30}},
                    height: 600,
                    width: 800,
                    mark: {type: "bar", tooltip: true},
                    encoding: {
                        y: {
                            field: "name",
                            type: "nominal",
                            axis: {title: "", labelColor: "white", labelFontSize: 15},
                            sort: "-x",
                        },
                        x: {
                            field: "count",
                            type: "quantitative",
                            axis: {
                                title: "Submissions",
                                titleColor: "white",
                                labelColor: "white",
                                labelFontSize: 15,
                            },
                        },
                    },
                }}/>

            <VegaLite
                data={{data: topCreditedPerformers}}
                spec={{
                    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
                    title: "Top Credited Performers",
                    background: "transparent",
                    data: {name: "data"},
                    config: {title: {color: "white", fontSize: 30}},
                    height: 600,
                    width: 800,
                    mark: {type: "bar", tooltip: true},
                    encoding: {
                        y: {
                            field: "name",
                            type: "nominal",
                            axis: {title: "", labelColor: "white", labelFontSize: 15},
                            sort: "-x",
                        },
                        x: {
                            field: "count",
                            type: "quantitative",
                            axis: {
                                title: "Submissions",
                                titleColor: "white",
                                labelColor: "white",
                                labelFontSize: 15,
                            },
                        },
                    },
                }}/>

            <VegaLite
                data={{data: topUniqueParts}}
                spec={{
                    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
                    title: "Most Unique Credited",
                    data: {name: "data"},
                    background: "transparent",
                    config: {title: {color: "white", fontSize: 30}},
                    height: 600,
                    width: 800,
                    mark: {type: "bar", tooltip: true},
                    encoding: {
                        y: {
                            field: "name",
                            type: "nominal",
                            axis: {title: "", labelColor: "white", labelFontSize: 15},
                            sort: "-x",
                        },
                        x: {
                            field: "count",
                            type: "quantitative",
                            axis: {
                                title: "Submissions",
                                titleColor: "white",
                                labelColor: "white",
                                labelFontSize: 15,
                            },
                        },
                    },
                }}/>
        </div>
    </div>;
};
export default Members;
