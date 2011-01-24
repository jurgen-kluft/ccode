using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Linq;

namespace Mercurial
{
    /// <summary>
    /// This class encapsulates a repository on disk.
    /// </summary>
    public sealed class Repository
    {
        private readonly string _Path;

        /// <summary>
        /// Initializes a new instance of the <see cref = "Repository" /> class.
        /// </summary>
        /// <param name = "rootPath">
        /// The path where the repository is stored locally, or the
        /// path to a directory that will be initialized with a new
        /// repository.
        /// </param>
        /// <exception cref="MercurialMissingException">
        /// The Mercurial command line client could not be located.
        /// </exception>
        /// <exception cref = "ArgumentNullException">
        /// <para><paramref name = "rootPath" /> is <c>null</c> or empty.</para>
        /// </exception>
        /// <exception cref = "DirectoryNotFoundException">
        /// <para><paramref name = "rootPath" /> refers to a directory that does not exist.</para>
        /// </exception>
        /// <exception cref = "InvalidOperationException">
        /// <para><paramref name = "rootPath" /> refers to a directory that doesn't appear to contain
        /// a Mercurial repository (no .hg directory.)</para>
        /// </exception>
        public Repository(string rootPath)
        {
            if (rootPath == null || rootPath.Trim().Length == 0)
                throw new ArgumentNullException("rootPath");
            if (!Directory.Exists(rootPath))
                throw new DirectoryNotFoundException("The specified directory for the Mercurial repository root does not exist");
            if (!Client.CouldLocateClient)
                throw new MercurialMissingException("The Mercurial command line client could not be located");

            _Path = System.IO.Path.GetFullPath(rootPath);
        }

        /// <summary>
        /// Gets the path of the repository root.
        /// </summary>
        /// <value>
        /// The path of the repository root.
        /// </value>
        public string Path
        {
            get
            {
                return _Path;
            }
        }

        /// <summary>
        /// Checks if the .hg folder exists
        /// </summary>
        public bool Exists
        {
            get
            {
                bool ok = Directory.Exists(_Path + "\\.hg");
                return ok;
            }
        }

        /// <summary>
        /// Executes the given <see cref="IMercurialCommand{TResult}"/> command against
        /// the Mercurial repository, returning the result as a typed value.
        /// </summary>
        /// <typeparam name="TResult">
        /// The type of result that is returned from executing the command.
        /// </typeparam>
        /// <param name="command">
        /// The <see cref="IMercurialCommand{T}"/> command to execute.
        /// </param>
        /// <returns>
        /// The result of executing the command.
        /// </returns>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="command"/> is <c>null</c>.</para>
        /// </exception>
        public TResult Execute<TResult>(IMercurialCommand<TResult> command)
        {
            if (command == null)
                throw new ArgumentNullException("command");

            Client.Execute(_Path, command);
            return command.Result;
        }

        /// <summary>
        /// Executes the given <see cref="IMercurialCommand"/> command against
        /// the Mercurial repository.
        /// </summary>
        /// <param name="command">
        /// The <see cref="IMercurialCommand"/> command to execute.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="command"/> is <c>null</c>.</para>
        /// </exception>
        public void Execute(IMercurialCommand command)
        {
            if (command == null)
                throw new ArgumentNullException("command");

            Client.Execute(_Path, command);
        }

        /// <summary>
        /// Gets all the changesets in the log.
        /// </summary>
        /// <param name="command">
        /// Any extra options to the log method, or <c>null</c> for default options.
        /// </param>
        /// <returns>
        /// A collection of <see cref = "Changeset" /> instances.
        /// </returns>
        public IEnumerable<Changeset> Log(LogCommand command = null)
        {
            command = command ?? new LogCommand();
            return Execute(command);
        }

        /// <summary>
        /// Gets all the changesets in the log.
        /// </summary>
        /// <param name="set">
        /// The <see cref="RevSpec"/> that specifies the set of revisions
        /// to include in the log. If <c>null</c>, return the whole log.
        /// Default is <c>null</c>.
        /// </param>
        /// <param name="command">
        /// Any extra options to the log method, or <c>null</c> for default options.
        /// </param>
        /// <returns>
        /// A collection of <see cref="Changeset" /> instances.
        /// </returns>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="set"/> is <c>null</c>.</para>
        /// </exception>
        /// <exception cref="ArgumentException">
        /// <para><paramref name="command"/>.<see cref="LogCommand.Revisions"/> is non-<c>null</c>, this is not valid for this method.</para>
        /// </exception>
        public IEnumerable<Changeset> Log(RevSpec set, LogCommand command = null)
        {
            if (set == null)
                throw new ArgumentNullException("set");
            if (command != null && command.Revisions != null)
                throw new ArgumentException("LogOptions.Revision cannot be set before calling this method");

            command = command ?? new LogCommand();
            command.Revisions.Add(set);
            return Execute(command);
        }

        /// <summary>
        /// Initializes a new repository in the directory this <see cref="Repository"/> refers to.
        /// </summary>
        /// <param name="command">
        /// Any extra options to the init method, or <c>null</c> for default options.
        /// </param>
        /// <exception cref="MercurialExecutionException">
        /// </exception>
        /// <exception cref="MercurialException">
        /// </exception>
        public void Init(InitCommand command = null)
        {
            command = command ?? new InitCommand();
            Execute(command);
        }

        /// <summary>
        /// Commits the specified files or all outstanding changes to the repository.
        /// </summary>
        /// <param name="command">
        /// Any extra options to the commit method.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="command"/> is <c>null</c>.</para>
        /// <para>- or -</para>
        /// <para><paramref name="command"/>.<see cref="CommitCommand.Message">Message</see> is empty.</para>
        /// </exception>
        public void Commit(CommitCommand command)
        {
            if (command == null)
                throw new ArgumentNullException("command");
            if (StringEx.IsNullOrWhiteSpace(command.Message))
                throw new ArgumentNullException("command", "The Message property of the commit options cannot be an empty string");

            Execute(command);
        }

        /// <summary>
        /// Commits the specified files or all outstanding changes to the repository.
        /// </summary>
        /// <param name="message">
        /// The commit message to use.
        /// </param>
        /// <param name="command">
        /// Any extra options to the commit method, or <c>null</c> for default options.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="message"/> is <c>null</c> or empty.</para>
        /// </exception>
        public void Commit(string message, CommitCommand command = null)
        {
            if (StringEx.IsNullOrWhiteSpace(message))
                throw new ArgumentNullException("message");

            command = command ?? new CommitCommand();
            command.Message = message;
            Execute(command);
        }

        /// <summary>
        /// Updates the working copy to a new revision.
        /// </summary>
        /// <param name="command">
        /// Any extra options to the update method, or <c>null</c> for default options.
        /// </param>
        public void Update(UpdateCommand command = null)
        {
            Execute(command ?? new UpdateCommand());
        }

        /// <summary>
        /// Updates the working copy to a new revision.
        /// </summary>
        /// <param name="revision">
        /// The revision to update the working copy to.
        /// </param>
        /// <param name="command">
        /// Any extra options to the update method, or <c>null</c> for default options.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="revision"/> is <c>null</c>.</para>
        /// </exception>
        public void Update(RevSpec revision, UpdateCommand command = null)
        {
            if (revision == null)
                throw new ArgumentNullException("revision");

            command = command ?? new UpdateCommand();
            command.Revision = revision;
            Execute(command);
        }

        /// <summary>
        /// Retrieves the status of changed files in the working directory.
        /// </summary>
        /// <returns>
        /// A collection of <see cref="FileStatus"/> objects, one for each modified file.
        /// </returns>
        public IEnumerable<FileStatus> Status(StatusCommand command = null)
        {
            command = command ?? new StatusCommand();
            return Execute(command);
        }

        /// <summary>
        /// Clones a repository into this <see cref="Repository"/>.
        /// </summary>
        /// <param name="command">
        /// Any extra options to the clone method.
        /// </param>
        public void Clone(CloneCommand command)
        {
            if (command == null)
                throw new ArgumentNullException("command");

            Execute(command);
        }

        /// <summary>
        /// 
        /// </summary>
        /// <param name="source">
        /// The path or Uri to the source to clone from.
        /// </param>
        /// <param name="command">
        /// Any extra options to the clone method, or <c>null</c> for default options.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="source"/> is <c>null</c> or empty.</para>
        /// </exception>
        /// <exception cref="ArgumentException">
        /// <para><see cref="CloneCommand.Source"/> cannot be set before calling this method.</para>
        /// </exception>
        public void Clone(string source, CloneCommand command = null)
        {
            if (StringEx.IsNullOrWhiteSpace(source))
                throw new ArgumentNullException("source");
            if (command != null && !StringEx.IsNullOrWhiteSpace(command.Source))
                throw new ArgumentException("CloneCommand.Source cannot be set before calling this method");

            command = command ?? new CloneCommand();
            command.Source = source;
            Execute(command);
        }

        /// <summary>
        /// Add all new files, delete all missing files.
        /// </summary>
        /// <param name="command">
        /// Any extra options to the addremove method, or <c>null</c> for default options.
        /// </param>
        public void AddRemove(AddRemoveCommand command = null)
        {
            command = command ?? new AddRemoveCommand();
            Execute(command);
        }

        /// <summary>
        /// Pull changes from the specified source.
        /// </summary>
        /// <param name="command">
        /// Any extra options to the pull method, or <c>null</c> for default options.
        /// </param>
        public void Pull(PullCommand command = null)
        {
            command = command ?? new PullCommand();
            Execute(command);
        }

        /// <summary>
        /// Pull changes from the specified source.
        /// </summary>
        /// <param name="source">
        /// The name of the source or the URL to the source, to pull from.
        /// </param>
        /// <param name="command">
        /// Any extra options to the pull method, or <c>null</c> for default options.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="source"/> is <c>null</c> or empty.</para>
        /// </exception>
        /// <exception cref="ArgumentException">
        /// <para><see cref="PullCommand.Source"/> cannot be set before calling this method.</para>
        /// </exception>
        public void Pull(string source, PullCommand command = null)
        {
            if (StringEx.IsNullOrWhiteSpace(source))
                throw new ArgumentNullException("source");
            if (command != null && !StringEx.IsNullOrWhiteSpace(command.Source))
                throw new ArgumentException("PullCommand.Source cannot be set before calling this method");

            command = command ?? new PullCommand();
            command.Source = source;
            Execute(command);
        }

        /// <summary>
        /// Push changes to the specified destination.
        /// </summary>
        /// <param name="command">
        /// Any extra options to the push method, or <c>null</c> for default options.
        /// </param>
        public void Push(PushCommand command = null)
        {
            command = command ?? new PushCommand();
            Execute(command);
        }

        /// <summary>
        /// Push changes to the specified destination.
        /// </summary>
        /// <param name="destination">
        /// The name of the destination or the URL to the destination, to push to.
        /// </param>
        /// <param name="command">
        /// Any extra options to the push method, or <c>null</c> for default options.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="destination"/> is <c>null</c> or empty.</para>
        /// </exception>
        /// <exception cref="ArgumentException">
        /// <para><see cref="PushCommand.Destination"/> cannot be set before calling this method.</para>
        /// </exception>
        public void Push(string destination, PushCommand command = null)
        {
            if (StringEx.IsNullOrWhiteSpace(destination))
                throw new ArgumentNullException("destination");
            if (command != null && !StringEx.IsNullOrWhiteSpace(command.Destination))
                throw new ArgumentException("PushCommand.Destination cannot be set before calling this method", "command");

            command = command ?? new PushCommand();
            command.Destination = destination;
            Execute(command);
        }

        /// <summary>
        /// Get current repository heads or get branch heads.
        /// </summary>
        /// <param name="command">
        /// Any extra options to the heads method, or <c>null</c> for default options.
        /// </param>
        /// <returns>
        /// A collection of <see cref="Changeset" /> instances.
        /// </returns>
        public IEnumerable<Changeset> Heads(HeadsCommand command = null)
        {
            command = command ?? new HeadsCommand();
            return Execute(command);
        }

        /// <summary>
        /// Annotates the specified item, returning annotation objects for the lines of
        /// the file.
        /// </summary>
        /// <param name="command">
        /// Any extra options to the annotate method, including the
        /// path to the item to annotate.
        /// </param>
        /// <returns>
        /// A collection of <see cref="Annotation"/> objects, one for each line.
        /// </returns>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="command"/> is <c>null</c>.</para>
        /// </exception>
        /// <exception cref="ArgumentException">
        /// <para><see cref="AnnotateCommand.Path"/> cannot be set before calling this method.</para>
        /// </exception>
        public IEnumerable<Annotation> Annotate(AnnotateCommand command = null)
        {
            if (command == null)
                throw new ArgumentNullException("command");
            if (StringEx.IsNullOrWhiteSpace(command.Path))
                throw new ArgumentException("AnnotateCommand.Path must be set before calling this method", "command");

            return Execute(command);
        }

        /// <summary>
        /// Annotates the specified item, returning annotation objects for the lines of
        /// the file.
        /// </summary>
        /// <param name="path">
        /// The path to the item to annotate.
        /// </param>
        /// <param name="command">
        /// Any extra options to the annotate method, or <c>null</c> for default options.
        /// </param>
        /// <returns>
        /// A collection of <see cref="Annotation"/> objects, one for each line.
        /// </returns>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="path"/> is <c>null</c> or empty.</para>
        /// </exception>
        /// <exception cref="ArgumentException">
        /// <para><see cref="AnnotateCommand.Path"/> cannot be set before calling this method.</para>
        /// </exception>
        public IEnumerable<Annotation> Annotate(string path, AnnotateCommand command = null)
        {
            if (StringEx.IsNullOrWhiteSpace(path))
                throw new ArgumentNullException("path");
            if (command != null && !StringEx.IsNullOrWhiteSpace(command.Path))
                throw new ArgumentException("AnnotateCommand.Path cannot be set before calling this method", "command");

            command = command ?? new AnnotateCommand();
            command.Path = path;
            return Execute(command);
        }

        /// <summary>
        /// Gets or sets the current branch name.
        /// </summary>
        /// <param name="command">
        /// Any extra options to the branch method, or <c>null</c> for default options.
        /// </param>
        /// <returns>
        /// The current or new branch name.
        /// </returns>
        public string Branch(BranchCommand command = null)
        {
            command = command ?? new BranchCommand();
            return Execute(command);
        }

        public bool HasOutstandingChanges
        {
            get
            {
                Mercurial.StatusCommand hg_status = new StatusCommand();
                hg_status.Include = FileStatusIncludes.Tracked;
                IEnumerable<FileStatus> status = Status(hg_status);
                foreach (FileStatus fs in status)
                    return true;
                return false;
            }
        }

        public Changeset GetChangeSet()
        {
            LogCommand hg_log = new LogCommand();
            hg_log.AddArgument("-l 1");
            IEnumerable<Changeset> hg_changesets = Log(hg_log);
            foreach (Changeset c in hg_changesets)
                return c;
            return null;
        }

        /// <summary>
        /// Gets or sets the current branch name.
        /// </summary>
        /// <param name="name">
        /// The name to use for the new branch.
        /// </param>
        /// <param name="command">
        /// Any extra options to the branch method, or <c>null</c> for default options.
        /// </param>
        /// <returns>
        /// The current or new branch name.
        /// </returns>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="name"/> is <c>null</c> or empty.</para>
        /// </exception>
        /// <exception cref="ArgumentException">
        /// <para><see cref="BranchCommand.Name"/> cannot be set before calling this method.</para>
        /// </exception>
        public string Branch(string name, BranchCommand command = null)
        {
            if (StringEx.IsNullOrWhiteSpace(name))
                throw new ArgumentNullException("name");
            if (command != null && !StringEx.IsNullOrWhiteSpace(command.Name))
                throw new ArgumentException("BranchCommand.Name cannot be set before calling this method", "command");

            command = command ?? new BranchCommand();
            command.Name = name;
            return Execute(command);
        }

        /// <summary>
        /// Retrieve aliases for remote repositories.
        /// </summary>
        /// <param name="command">
        /// Any extra options to the paths method, or <c>null</c> for default options.
        /// </param>
        /// <returns>
        /// A collection of <see cref="RemoteRepositoryPath"/> objects, one for each
        /// remote repository.
        /// </returns>
        public IEnumerable<RemoteRepositoryPath> Paths(PathsCommand command = null)
        {
            command = command ?? new PathsCommand();
            return Execute(command);
        }

        /// <summary>
        /// Adds one or more files to the repository.
        /// </summary>
        /// <param name="path">
        /// The path to a file, or a path containing wildcards to files to add
        /// to the repository.
        /// </param>
        /// <param name="command">
        /// The information object for the add command, containing the paths of the files
        /// to add.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="path"/> is <c>null</c> or empty.</para>
        /// </exception>
        public void Add(string path, AddCommand command = null)
        {
            if (StringEx.IsNullOrWhiteSpace(path))
                throw new ArgumentNullException("path");

            command = command ?? new AddCommand();
            command.Paths.Add(path);
            Execute(command);
        }

        /// <summary>
        /// Adds one or more files to the repository.
        /// </summary>
        /// <param name="command">
        /// The information object for the add command, containing the paths of the files
        /// to add.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="command"/> is <c>null</c>.</para>
        /// </exception>
        public void Add(AddCommand command)
        {
            if (command == null)
                throw new ArgumentNullException("command");

            Execute(command);
        }

        /// <summary>
        /// Removes one or more files from the repository.
        /// </summary>
        /// <param name="path">
        /// The path to a file, or a path containing wildcards to files to remove
        /// from the repository.
        /// </param>
        /// <param name="command">
        /// The information object for the remove command, containing the paths of the files
        /// to remove.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="path"/> is <c>null</c> or empty.</para>
        /// </exception>
        public void Remove(string path, RemoveCommand command = null)
        {
            if (StringEx.IsNullOrWhiteSpace(path))
                throw new ArgumentNullException("path");

            command = command ?? new RemoveCommand();
            command.Paths.Add(path);
            Execute(command);
        }

        /// <summary>
        /// Removes one or more files from the repository.
        /// </summary>
        /// <param name="command">
        /// The information object for the remove command, containing the paths of the files
        /// to remove.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="command"/> is <c>null</c>.</para>
        /// </exception>
        public void Remove(RemoveCommand command)
        {
            if (command == null)
                throw new ArgumentNullException("command");

            Execute(command);
        }

        /// <summary>
        /// Forgets one or more tracked files in the repository, making Mercurial
        /// stop tracking them.
        /// </summary>
        /// <param name="path">
        /// The path to a file, or a path containing wildcards to files to forget
        /// in the repository.
        /// </param>
        /// <param name="command">
        /// The information object for the forget command, containing the paths of the files
        /// to forget.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="path"/> is <c>null</c> or empty.</para>
        /// </exception>
        public void Forget(string path, ForgetCommand command = null)
        {
            if (StringEx.IsNullOrWhiteSpace(path))
                throw new ArgumentNullException("path");

            command = command ?? new ForgetCommand();
            command.Paths.Add(path);
            Execute(command);
        }

        /// <summary>
        /// Forgets one or more tracked files in the repository, making Mercurial
        /// stop tracking them.
        /// </summary>
        /// <param name="command">
        /// The information object for the forget command, containing the paths of the files
        /// to forget.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="command"/> is <c>null</c>.</para>
        /// </exception>
        public void Forget(ForgetCommand command)
        {
            if (command == null)
                throw new ArgumentNullException("command");

            Execute(command);
        }

        /// <summary>
        /// Retrieves the tip revision.
        /// </summary>
        /// <param name="command">
        /// The information object for the tip command.
        /// </param>
        /// <returns>
        /// The <see cref="Changeset"/> of the tip revision.
        /// </returns>
        public Changeset Tip(TipCommand command = null)
        {
            command = command ?? new TipCommand();
            return Execute(command);
        }

        /// <summary>
        /// Retrieves the tags.
        /// </summary>
        /// <param name="command">
        /// The information object for the tags command.
        /// </param>
        /// <returns>
        /// The <see cref="text"/> of the tags.
        /// </returns>
        public IEnumerable<Tag> Tags(TagsCommand command = null)
        {
            command = command ?? new TagsCommand();
            return Execute(command);
        }

        /// <summary>
        /// Add or remove a tag for a changeset.
        /// </summary>
        /// <param name="name">
        /// The name of the tag.
        /// </param>
        /// <param name="command">
        /// The information object for the tag command, or <c>null</c> if no extra information
        /// is necessary.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="name"/> is <c>null</c> or empty.</para>
        /// </exception>
        /// <exception cref="InvalidOperationException">
        /// <para><paramref name="command"/>.<see cref="TagCommand.Name">Name</see> and <paramref name="name"/> was both set.</para>
        /// </exception>
        public void Tag(string name, TagCommand command = null)
        {
            if (StringEx.IsNullOrWhiteSpace(name))
                throw new ArgumentNullException("name");
            if (command != null && !StringEx.IsNullOrWhiteSpace(command.Name))
                throw new InvalidOperationException("Both name and command.Name cannot be set before calling Tag");

            command = command ?? new TagCommand();
            command.Name = name;
            Execute(command);
        }

        /// <summary>
        /// Retrieve changesets not found in the default destination.
        /// </summary>
        /// <param name="command">
        /// Any extra options to the outgoing method, or <c>null</c> for default options.
        /// </param>
        public IEnumerable<Changeset> Outgoing(OutgoingCommand command = null)
        {
            command = command ?? new OutgoingCommand();
            Execute(command);
            return command.Result;
        }

        /// <summary>
        /// Retrieve changesets not found in the destination.
        /// </summary>
        /// <param name="destination">
        /// The name of the destination or the URL to the destination, to check outgoing to.
        /// </param>
        /// <param name="command">
        /// Any extra options to the outgoing method, or <c>null</c> for default options.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="destination"/> is <c>null</c> or empty.</para>
        /// </exception>
        /// <exception cref="ArgumentException">
        /// <para><see cref="OutgoingCommand.Destination"/> cannot be set before calling this method.</para>
        /// </exception>
        public IEnumerable<Changeset> Outgoing(string destination, OutgoingCommand command = null)
        {
            if (StringEx.IsNullOrWhiteSpace(destination))
                throw new ArgumentNullException("destination");
            if (command != null && !StringEx.IsNullOrWhiteSpace(command.Destination))
                throw new ArgumentException("OutgoingCommand.Destination cannot be set before calling this method", "command");

            command = command ?? new OutgoingCommand();
            command.Destination = destination;
            Execute(command);
            return command.Result;
        }

        /// <summary>
        /// Retrieve new changesets found in the default source.
        /// </summary>
        /// <param name="command">
        /// Any extra options to the incoming method, or <c>null</c> for default options.
        /// </param>
        public IEnumerable<Changeset> Incoming(IncomingCommand command = null)
        {
            command = command ?? new IncomingCommand();
            Execute(command);
            return command.Result;
        }

        /// <summary>
        /// Retrieve new changesets found in the source.
        /// </summary>
        /// <param name="source">
        /// The name of the source or the URL to the source, to check incoming from.
        /// </param>
        /// <param name="command">
        /// Any extra options to the incoming method, or <c>null</c> for default options.
        /// </param>
        /// <exception cref="ArgumentNullException">
        /// <para><paramref name="source"/> is <c>null</c> or empty.</para>
        /// </exception>
        /// <exception cref="ArgumentException">
        /// <para><see cref="IncomingCommand.Source"/> cannot be set before calling this method.</para>
        /// </exception>
        public IEnumerable<Changeset> Incoming(string source, IncomingCommand command = null)
        {
            if (StringEx.IsNullOrWhiteSpace(source))
                throw new ArgumentNullException("source");
            if (command != null && !StringEx.IsNullOrWhiteSpace(command.Source))
                throw new ArgumentException("IncomingCommand.Source cannot be set before calling this method", "command");

            command = command ?? new IncomingCommand();
            command.Source = source;
            Execute(command);
            return command.Result;
        }
    }
}