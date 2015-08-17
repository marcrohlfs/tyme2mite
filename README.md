# Tyme to Mite

Read time entries from [Tyme](http://tyme-app.com) and send them to [mite](http://mite.yo.lk).

**Please note**: I'm new to [Go](http://golang.org/) and thus not yet familiar with how to write good Go code. To get
the initial version of the _tyme2mite_ command running as fast as possible, I put all the code into one single function.
I'll now start to improve things little by little - when I find some time besides my professional work. See _Objectives_
for (slight) information about further implementation plans.

## Installation

Should be described on the [Go Documentation](http://golang.org/doc/) page ...

### Dependencies

This package depends on some non-bundled packages:

* _github.com/everdev/mack_
* _github.com/jimlawless/cfg_

They should normally be installed automatically - depending on who the user installs _this_ package.

### Configuration

The _tyme2mite_ command requires a configuration file named ``tyme2mite.cfg`` in the users' home directory. It must
contain the following configuration entries:

* ``mite_base_url``: The base URL of the _mite_ service. Example: ``https://DEMO.mite.yo.lk`` (``DEMO`` should be
  substituted with the name of Your company/organisation account).
* ``mite_api_key``: The API key for user authentication. It has to be activated and displayed in the user settings
  [https://DEMO.mite.yo.lk/myself](https://DEMO.mite.yo.lk/myself).
* ``mite_import_active``: Entries will only be sent to _mite_ if this configuration parameter is defined and its value
  is explicitly set to ``true``. This should prevent unwanted imports to productive _mite_ during test phase.

## Program Arguments

The _tyme2mite_ command can be executed with two program arguments:

1. The _start date_ of the _Tyme_ entries to be transferred to _mite_.
2. The _end date_ of the _Tyme_ entries to be transferred to _mite_.

If the second argument is omitted, the current date will be used as _end date_. If both arguments are omitted, all
entries for _yesterday_ and _today_ will be transferred.

The format of the _date_ arguments is ``yyyy-mm-dd``.

## Data Mapping between Tyme and Mite

As _customers_, _projects_ and _services_ are expected to be changed regularly - either in _mite_ or _Tyme_ -
``tyme2mite`` doesn't work with any mapping description file. Instead, the names of _projects_ and _tasks_ in _Tyme_
must match a pattern that allows to resolve the regarding _customers_, _projects_ and _services_ in _mite_ by their
names:

* The pattern for _project_ names is ``[CUSTOMER_NAME] | [PROJECT_NAME]``.
* There are two patterns for _task_ names: ``[SERVICE_NAME]`` and ``[CUSTOM_TASK_NAME] | [SERVICE_NAME]``. In latter
  case, the _custom taks name_ will be prepended before the note.

Time entries with the same _date_, _customer_, _project_, _service_ and _custom task name_ will merged into one entry
before they're sent to _mite_.

## Objectives

* Instead of requiring name patterns in _Tyme_, _tags_ may be assigned to _projects_ and _tasks_. These _tags_ would
  describe which _customers_, _projects_ and _services_ should be mapped in _mite_. They could look like follows:
    * ``mite:customer-name=The Customer`` (assigned to a project)
    * ``mite:project-name=The Poject`` (assigned to a project)
    * ``mite:service-name=A Service`` (assigned to tasks)
* Extract _Go_ library for _mite_
* Extract _Go_ library for _Tyme_
* Implement mappings for other tools

## References

API references used for implementation:

* [API. Documentation of mite's RESTful web service](http://mite.yo.lk/en/api/index.html)
* [Tyme Scripting | Tyme](http://tyme-app.com/tyme-scripting)

## Contributions

Contributions are highly appreciated.

**Please** be sure to use the _develop_ branch as parent for Your feature branches and as target for Your pull requests!
