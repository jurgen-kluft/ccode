using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Linq;
using System.Xml.Serialization;
using Mercurial.XmlSerializationTypes;

namespace Mercurial
{
    /// <summary>
    /// This class implements a basic XML-based changeset parser, that parses changeset information
    /// as reported by the Mercurial command line client, in XML format.
    /// </summary>
    public static class ChangesetXmlParser
    {
        /// <summary>
        /// Parse the given XML and return <see cref="Changeset"/> objects for the information
        /// contained in it.
        /// </summary>
        /// <param name="xml">
        /// The XML to parse.
        /// </param>
        /// <returns>
        /// An array of <see cref="Changeset"/> objects, or an empty array if no
        /// changeset is present (empty string most likely.)
        /// </returns>
        public static Changeset[] Parse(string xml)
        {
            if (StringEx.IsNullOrWhiteSpace(xml))
                return new Changeset[0];

            var serializer = new XmlSerializer(typeof (LogNode));
            var log = (LogNode) serializer.Deserialize(new StringReader(xml));

            var entryList = (from logEntry in log.LogEntries
                             select
                                 new
                                     {
                                         parents = logEntry.Parents,
                                         actions = logEntry.PathActions,
                                         changeset =
                                 new Changeset
                                     {
                                         Timestamp = logEntry.Timestamp,
                                         AuthorName = (logEntry.Author ?? new LogEntryAuthorNode()).Name,
                                         AuthorEmailAddress = (logEntry.Author ?? new LogEntryAuthorNode()).Email,
                                         CommitMessage = logEntry.CommitMessage ?? string.Empty,
                                         Branch = logEntry.Branch ?? "default",
                                         Hash = logEntry.Hash,
                                         RevisionNumber = logEntry.Revision,
                                         Revision = RevSpec.Single(logEntry.Hash),
                                         Tags = logEntry.Tags.Select(t => t.Name).ToArray(),
                                     }
                                     }).ToList();

            foreach (var entry in entryList)
            {
                foreach (LogEntryPathNode action in entry.actions)
                {
                    var pathAction = new ChangesetPathAction { Path = action.Path, };
                    switch (action.Action)
                    {
                        case "M":
                            pathAction.Action = ChangesetPathActionType.Modify;
                            break;

                        case "A":
                            pathAction.Action = ChangesetPathActionType.Add;
                            break;

                        case "R":
                            pathAction.Action = ChangesetPathActionType.Remove;
                            break;

                        default:
                            throw new InvalidOperationException("Unknown path action: " + action.Action);
                    }
                    entry.changeset.PathActions.Add(pathAction);
                }
            }

            Dictionary<int, string> lookup = entryList.ToDictionary(e => e.changeset.RevisionNumber, e => e.changeset.Hash);

            foreach (var entry in entryList)
            {
                switch (entry.parents.Count)
                {
                    case 0:
                        if (entry.changeset.RevisionNumber > 0)
                        {
                            if (lookup.ContainsKey(entry.changeset.RevisionNumber - 1))
                                entry.changeset.LeftParentHash = lookup[entry.changeset.RevisionNumber - 1];
                            else
                                entry.changeset.LeftParentHash = String.Empty;
                            entry.changeset.LeftParentRevision = entry.changeset.RevisionNumber - 1;
                        }
                        else
                        {
                            entry.changeset.LeftParentHash = String.Empty;
                            entry.changeset.LeftParentRevision = -1;
                        }

                        entry.changeset.RightParentHash = String.Empty;
                        entry.changeset.RightParentRevision = -1;
                        break;

                    case 1:
                        entry.changeset.LeftParentHash = entry.parents[0].Hash;
                        entry.changeset.LeftParentRevision = entry.parents[0].Revision;

                        entry.changeset.RightParentHash = String.Empty;
                        entry.changeset.RightParentRevision = -1;
                        break;

                    case 2:
                        entry.changeset.LeftParentHash = entry.parents[0].Hash;
                        entry.changeset.LeftParentRevision = entry.parents[0].Revision;

                        entry.changeset.RightParentHash = entry.parents[1].Hash;
                        entry.changeset.RightParentRevision = entry.parents[1].Revision;
                        break;

                    default:
                        throw new InvalidOperationException("Invalid number of parents for changeset " + entry.changeset.Hash + ", has " +
                                                            entry.parents.Count + " parents");
                }
            }

            return (from entry in entryList
                    orderby entry.changeset.RevisionNumber descending
                    select entry.changeset).ToArray();
        }
    }
}