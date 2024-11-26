import { Checkbox, Input, List, ListItem, VStack } from "@chakra-ui/react";
import { debounce } from "lodash";
import { useState, useEffect } from "react";

const GrantCheckList = ({ grantOptions, onChange, parentStyles }) => {
    const [allGroups, setAllGroups] = useState(false);
    const [selectedGroups, setSelectedGroups] = useState([]);
    const [grants, setGrants] = useState([]);
    const [searchTerm, setSearchTerm] = useState(""); // State to manage the search query

    // Group the grants by their prefix (category)
    const groupGrants = (grants) => {
        const groupMap = new Map();

        grants.forEach(grant => {
            const [prefix] = grant.grant.split('-');

            if (!groupMap.has(prefix)) {
                groupMap.set(prefix, []);
            }

            groupMap.get(prefix).push(grant);
        });

        return Array.from(groupMap, ([group, grants]) => ({ group, grants }));
    };

    useEffect(() => {
        setGrants(groupGrants(grantOptions));
    }, [grantOptions]);

    // Function to handle the search input change
    const handleSearch = (e) => {
        setSearchTerm(e.target.value);
    };

    // Filter grants based on the search term
    const filteredGrants = grants.map(group => ({
        ...group,
        grants: group.grants.filter(grant =>
            grant.grant.toLowerCase().includes(searchTerm.toLowerCase()) // Case-insensitive search
        )
    })).filter(group => group.grants.length > 0); // Only include groups that have matching grants

    // Debounced update handler for selected grants
    const handleUpdate = debounce((selected) => {
        onChange(selected); // Send the latest selected values to parent component
    }, 500);

    useEffect(() => {
        const selected = [
            ...selectedGroups, // Add selected group names
            ...filteredGrants.flatMap(group =>
                group.grants.filter(grant => grant.selected && !selectedGroups.includes(group.group)).map(grant => grant.grant)
            ),
        ];
        handleUpdate(selected); // Replace with actual save logic
    }, [filteredGrants, selectedGroups]);

    // Handle group checkbox changes
    const handleAllChange = (isChecked) => {
        let updatedGroups = [];

        // Update individual grants based on the group selection
        const updatedGrants = grants.map(group => {
            if (isChecked) {
                updatedGroups.push(group.group)
            }
            return {
                ...group,
                grants: group.grants.map(grant => ({ ...grant, selected: isChecked }))
            };
        });
        setSelectedGroups(updatedGroups);
        setGrants(updatedGrants);
        setAllGroups(isChecked)
    };

    // Handle group checkbox changes
    const handleGroupChange = (groupName, isChecked) => {
        let updatedGroups = [...selectedGroups];
        if (isChecked) {
            if (!updatedGroups.includes(groupName)) updatedGroups.push(groupName);
        } else {
            updatedGroups = updatedGroups.filter(group => group !== groupName);
        }

        setSelectedGroups(updatedGroups);

        // Update individual grants based on the group selection
        const updatedGrants = grants.map(group => {
            if (group.group === groupName) {
                return {
                    ...group,
                    grants: group.grants.map(grant => ({ ...grant, selected: isChecked }))
                };
            }
            return group;
        });

        setGrants(updatedGrants);
    };

    // Handle individual grant checkbox changes
    const handleGrantChange = (groupName, grantName, isChecked) => {
        const updatedGrants = filteredGrants.map(group => {
            if (group.group === groupName) {
                const updatedGrant = group.grants.map(grant =>
                    grant.grant === grantName ? { ...grant, selected: isChecked } : grant
                );
                return { ...group, grants: updatedGrant };
            }
            return group;
        });
        setGrants(updatedGrants);

        // Check if all grants in the group are selected to update the group checkbox
        const updatedGroups = filteredGrants.map(group => {
            if (group.group === groupName) {
                const allSelected = group.grants.every(grant => grant.selected);
                if (allSelected) {
                    // If all are selected, mark the group as selected
                    if (!selectedGroups.includes(groupName)) {
                        return { ...group, selected: true };  // Ensure state update
                    }
                } else {
                    // If not all are selected, uncheck the group
                    return { ...group, selected: false };  // Ensure state update
                }
            }
            return group;
        });

        setSelectedGroups(updatedGroups.map(group => group.group)); // Correctly update the selected groups
    };

    return (
        <VStack className={parentStyles.aclContainer}>
            {/* Search Input */}
            <Input
                id="searchAcl"
                type="search"
                onChange={handleSearch}
                placeholder="Search ACL"
            />

            {/* List of Grants */}
            <List className={parentStyles.aclList}>
                <ListItem key="allGroup" className={parentStyles.aclListItem}>
                    <Checkbox
                        isChecked={allGroups}
                        isIndeterminate={filteredGrants.some(group => group.grants.some(grant => grant.selected)) && !filteredGrants.every(group => group.grants.every(grant => grant.selected))}
                        onChange={(e) => handleAllChange(e.target.checked)}
                    >
                        Select All
                    </Checkbox>
                </ListItem>
                {filteredGrants.map((group) => (
                    <ListItem key={group.group} className={parentStyles.aclCategoryItem}>
                        <Checkbox
                            isChecked={selectedGroups.includes(group.group)}
                            isIndeterminate={group.grants.some(grant => grant.selected) && !group.grants.every(grant => grant.selected)}
                            onChange={(e) => handleGroupChange(group.group, e.target.checked)}
                        >
                            {group.group} {/* Group Name */}
                        </Checkbox>
                        <List pl={4} mt={2}>
                            {group.grants.map((grant) => (
                                <ListItem key={grant.grant} className={parentStyles.aclListItem}>
                                    <Checkbox
                                        key={grant.grant}
                                        isChecked={grant.selected}
                                        onChange={(e) => handleGrantChange(group.group, grant.grant, e.target.checked)}
                                    >
                                        {grant.grant} {/* Individual Grant */}
                                    </Checkbox>
                                </ListItem>
                            ))}
                        </List>
                    </ListItem>
                ))}
            </List>
        </VStack>
    );
};

export default GrantCheckList;
