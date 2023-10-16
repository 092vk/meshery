import React, { useEffect, useMemo, useRef, useState } from 'react';
import {
  NoSsr,
  TableCell,
  Button,
  Tooltip,
  Link,
  FormControl,
  Select,
  MenuItem,
  TableContainer,
  Table,
  Grid,
  TableRow,
  TableSortLabel,
  Chip,
} from '@material-ui/core';
import { withStyles } from '@material-ui/core/styles';
import DeleteForeverIcon from '@mui/icons-material/DeleteForever';
import Moment from 'react-moment';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import { updateProgress } from '../../lib/store';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import AssignmentTurnedInIcon from '@mui/icons-material/AssignmentTurnedIn';
import ExploreIcon from '@mui/icons-material/Explore';
import RemoveCircleIcon from '@mui/icons-material/RemoveCircle';
import classNames from 'classnames';
import dataFetch from '../../lib/data-fetch';
import LaunchIcon from '@mui/icons-material/Launch';
import { useNotification } from '../../utils/hooks/useNotification';
import { EVENT_TYPES } from '../../lib/event-types';
import CustomColumnVisibilityControl from '../../utils/custom-column';
import SearchBar from '../../utils/custom-search';
import ResponsiveDataTable from '../../utils/data-table';
import useStyles from '../../assets/styles/general/tool.styles';
import Modal from '../Modal';
import { Colors } from '../../themes/app';
import { iconMedium } from '../../css/icons.styles';
import PromptComponent from '../PromptComponent';
import { FormattedMetadata } from '../NotificationCenter/metadata';

const styles = (theme) => ({
  grid: { padding: theme.spacing(2) },
  tableHeader: {
    fontWeight: 'bolder',
    fontSize: 18,
  },
  muiRow: {
    '& .MuiTableCell-root': {
      textTransform: 'capitalize',
    },
  },
  statusSelect: {
    '& .MuiSelect-select.MuiSelect-select': {
      padding: '0 2px',
    },
    '& .MuiSelect-icon': {
      display: 'none',
    },
  },
  createButton: {
    display: 'flex',
    justifyContent: 'flex-start',
    alignItems: 'center',
    whiteSpace: 'nowrap',
  },
  viewSwitchButton: {
    justifySelf: 'flex-end',
    marginLeft: 'auto',
    paddingLeft: '1rem',
  },
  statusChip: {
    minWidth: '120px !important',
    maxWidth: 'max-content !important',
    display: 'flex !important',
    justifyContent: 'flex-start !important',
    textTransform: 'capitalize',
    borderRadius: '3px !important',
    padding: '6px 8px',
    '& .MuiChip-label': {
      paddingTop: '3px',
      fontWeight: '400',
    },
    '&:hover': {
      boxShadow: '0px 1px 2px 0px rgba(0, 0, 0, 0.25)',
      cursor: 'pointer',
    },
  },
  capitalize: {
    textTransform: 'capitalize',
  },
  lowecase: {
    textTransform: 'lowecase',
  },
  ignored: {
    '& .MuiChip-label': {
      color: `${theme.palette.secondary.default}`,
    },
    background: `${theme.palette.secondary.default}30 !important`,
    '& .MuiSvgIcon-root': {
      color: `${theme.palette.secondary.default} !important`,
    },
  },
  connected: {
    '& .MuiChip-label': {
      color: theme.palette.secondary.success,
    },
    background: `${theme.palette.secondary.success}30 !important`,
    '& .MuiSvgIcon-root': {
      color: `${theme.palette.secondary.success} !important`,
    },
  },
  registered: {
    '& .MuiChip-label': {
      color: theme.palette.secondary.primary,
    },
    background: `${theme.palette.secondary.primary}30 !important`,
    '& .MuiSvgIcon-root': {
      color: `${theme.palette.secondary.primary} !important`,
    },
  },
  discovered: {
    '& .MuiChip-label': {
      color: theme.palette.secondary.warning,
    },
    background: `${theme.palette.secondary.warning}30 !important`,
    '& .MuiSvgIcon-root': {
      color: `${theme.palette.secondary.warning} !important`,
    },
  },
  deleted: {
    '& .MuiChip-label': {
      color: theme.palette.secondary.error,
    },
    background: `${theme.palette.secondary.lightError}30 !important`,
    '& .MuiSvgIcon-root': {
      color: `${theme.palette.secondary.error} !important`,
    },
    '&:hover': {
      cursor: 'initial',
    },
  },
  expandedRows: {
    background: `${theme.palette.secondary.default}10`,
  },
  contentContainer: {
    [theme.breakpoints.down(1050)]: {
      flexDirection: 'column',
    },
    flexWrap: 'noWrap',
  },
  innerTableWrapper: {
    backgroundColor: '#cccccc',
    borderTop: '1px solid #E0E0E0',
    borderBottom: '1px solid #E0E0E0',
    borderRadius: 0,
    background: '#FFF',
    boxShadow:
      '0px 4px 8px 0px rgba(0, 0, 0, 0.08) inset, 0px 0px 4px 0px rgba(0, 0, 0, 0.25) inset, 1px 0px 4px 0px rgba(0, 0, 0, 0.25) inset, 0px -1px 4px 0px rgba(0, 0, 0, 0.25) inset',
    padding: '0',
  },
  innerTableContainer: {
    background: 'linear-gradient(90deg, #F2F0F0 0%, #FFF 100%)',
    margin: '1px',
    marginLeft: '18px',
    width: 'calc(100% - 20px)',
  },
  listItem: {
    paddingTop: '0',
    paddingBottom: '0',
  },
  noGutter: {
    padding: '0',
  },
  showMore: {
    color: Colors.keppelGreen,
    cursor: 'pointer',
  },
  bulkAction: {
    display: 'flex',
    justifyContent: 'flex-end',
    width: '100%',
  },
});

const ACTION_TYPES = {
  FETCH_CONNECTIONS: {
    name: 'FETCH_CONNECTIONS',
    error_msg: 'Failed to fetch connections',
  },
  UPDATE_CONNECTION: {
    name: 'UPDATE_CONNECTION',
    error_msg: 'Failed to update connection',
  },
  DELETE_CONNECTION: {
    name: 'DELETE_CONNECTION',
    error_msg: 'Failed to delete connection',
  },
};

/**
 * Parent Component for Connection Component
 *
 * @important
 * - Keep the component's responsibilities focused on connection management. Avoid adding unrelated functionality and state.
 */

function ConnectionManagementPage(props) {
  const [createConnectionModal, setCreateConnectionModal] = useState({
    open: false,
  });
  const [createConnection, setCreateConnection] = useState({});

  const handleCreateConnectionModalOpen = () => {
    setCreateConnectionModal({ open: true });
  };

  const handleCreateConnectionModalClose = () => {
    setCreateConnectionModal({ open: false });
  };

  const handleCreateConnectionSubmit = () => {};

  useEffect(() => {
    dataFetch(
      '/api/schema/resource/helmRepo',
      {
        method: 'GET',
        credentials: 'include',
      },
      (result) => {
        setCreateConnection(result);
      },
    );
  }, []);

  return (
    <>
      <Connections
        createConnectionModal={createConnectionModal}
        onOpenCreateConnectionModal={handleCreateConnectionModalOpen}
        onCloseCreateConnectionModal={handleCreateConnectionModalClose}
        {...props}
      />
      {createConnectionModal.open && (
        <Modal
          open={true}
          schema={createConnection.rjsfSchema}
          uiSchema={createConnection.uiSchema}
          handleClose={handleCreateConnectionModalClose}
          handleSubmit={handleCreateConnectionSubmit}
          title="Connect Helm Repository"
          submitBtnText="Connect"
          // leftHeaderIcon={ }
          // submitBtnIcon={<PublishIcon  className={classes.addIcon} data-cy="import-button"/>}
        />
      )}
    </>
  );
}

function Connections({ classes, updateProgress, onOpenCreateConnectionModal }) {
  const modalRef = useRef(null);
  const [page, setPage] = useState(0);
  const [count, setCount] = useState(0);
  const [pageSize, setPageSize] = useState(0);
  const [connections, setConnections] = useState([]);
  const [search, setSearch] = useState('');
  const [sortOrder, setSortOrder] = useState('');
  const [showMore, setShowMore] = useState(false);
  const [rowsExpanded, setRowsExpanded] = useState([]);
  const [loading, setLoading] = useState(false);

  const { notify } = useNotification();
  const StyleClass = useStyles();

  const statuses = ['ignored', 'connected', 'REGISTERED', 'discovered', 'deleted'];

  const status = (value) => {
    switch (value) {
      case 'ignored':
        return (
          <MenuItem value={value}>
            <Chip
              className={classNames(classes.statusChip, classes.ignored)}
              avatar={<RemoveCircleIcon />}
              label={value}
            />
          </MenuItem>
        );
      case 'connected':
        return (
          <MenuItem value={value}>
            <Chip
              className={classNames(classes.statusChip, classes.connected)}
              value={value}
              avatar={<CheckCircleIcon />}
              label={value}
            />
          </MenuItem>
        );
      case 'REGISTERED':
        return (
          <MenuItem value={value}>
            <Chip
              className={classNames(classes.statusChip, classes.registered)}
              value={value}
              avatar={<AssignmentTurnedInIcon />}
              label={value.toLowerCase()}
            />
          </MenuItem>
        );
      case 'discovered':
        return (
          <MenuItem value={value}>
            <Chip
              className={classNames(classes.statusChip, classes.discovered)}
              value={value}
              avatar={<ExploreIcon />}
              label={value}
            />
          </MenuItem>
        );
      case 'deleted':
        return (
          <MenuItem value={value}>
            <Chip
              className={classNames(classes.statusChip, classes.deleted)}
              value={value}
              avatar={<DeleteForeverIcon />}
              label={value}
            />
          </MenuItem>
        );
      default:
        return (
          <MenuItem value={value}>
            <Chip
              className={classNames(classes.statusChip, classes.discovered)}
              value={value}
              avatar={<ExploreIcon />}
              label={value}
            />
          </MenuItem>
        );
    }
  };

  const columns = [
    {
      name: 'id',
      label: 'ID',
      options: {
        display: false,
      },
    },
    {
      name: 'metadata.server_location',
      label: 'Server Location',
      options: {
        display: false,
      },
    },
    {
      name: 'metadata.server',
      label: 'Server Location',
      options: {
        display: false,
      },
    },
    {
      name: 'name',
      label: 'Name',
      options: {
        sort: true,
        sortThirdClickReset: true,
        customHeadRender: function CustomHead({ index, ...column }, sortColumn, columnMeta) {
          return (
            <SortableTableCell
              index={index}
              columnData={column}
              columnMeta={columnMeta}
              onSort={() => sortColumn(index)}
            />
          );
        },
        customBodyRender: (value, tableMeta) => {
          return (
            <Tooltip title={tableMeta.rowData[1] || tableMeta.rowData[2]} placement="top">
              <Link href={tableMeta.rowData[1] || tableMeta.rowData[2]} target="_blank">
                {value}
                <sup>
                  <LaunchIcon sx={{ fontSize: '12px' }} />
                </sup>
              </Link>
            </Tooltip>
          );
        },
      },
    },
    {
      name: 'kind',
      label: 'Kind',
      options: {
        sort: true,
        sortThirdClickReset: true,
        customHeadRender: function CustomHead({ index, ...column }, sortColumn, columnMeta) {
          return (
            <SortableTableCell
              index={index}
              columnData={column}
              columnMeta={columnMeta}
              onSort={() => sortColumn(index)}
            />
          );
        },
      },
    },
    {
      name: 'type',
      label: 'Category',
      options: {
        sort: true,
        sortThirdClickReset: true,
        customHeadRender: function CustomHead({ index, ...column }, sortColumn, columnMeta) {
          return (
            <SortableTableCell
              index={index}
              columnData={column}
              columnMeta={columnMeta}
              onSort={() => sortColumn(index)}
            />
          );
        },
      },
    },
    {
      name: 'sub_type',
      label: 'Sub Category',
      options: {
        sort: true,
        sortThirdClickReset: true,
        customHeadRender: function CustomHead({ index, ...column }, sortColumn, columnMeta) {
          return (
            <SortableTableCell
              index={index}
              columnData={column}
              columnMeta={columnMeta}
              onSort={() => sortColumn(index)}
            />
          );
        },
      },
    },
    {
      name: 'updated_at',
      label: 'Updated At',
      options: {
        sort: true,
        sortThirdClickReset: true,
        customHeadRender: function CustomHead({ index, ...column }, sortColumn, columnMeta) {
          return (
            <SortableTableCell
              index={index}
              columnData={column}
              columnMeta={columnMeta}
              onSort={() => sortColumn(index)}
            />
          );
        },
        customBodyRender: function CustomBody(value) {
          return (
            <Tooltip
              title={
                <Moment startOf="day" format="LLL">
                  {value}
                </Moment>
              }
              placement="top"
              arrow
              interactive
            >
              <Moment format="LL">{value}</Moment>
            </Tooltip>
          );
        },
      },
    },
    {
      name: 'created_at',
      label: 'Discovered At',
      options: {
        sort: true,
        sortThirdClickReset: true,
        customHeadRender: function CustomHead({ index, ...column }, sortColumn, columnMeta) {
          return (
            <SortableTableCell
              index={index}
              columnData={column}
              columnMeta={columnMeta}
              onSort={() => sortColumn(index)}
            />
          );
        },
        customBodyRender: function CustomBody(value) {
          return (
            <Tooltip
              title={
                <Moment startOf="day" format="LLL">
                  {value}
                </Moment>
              }
              placement="top"
              arrow
              interactive
            >
              <Moment format="LL">{value}</Moment>
            </Tooltip>
          );
        },
      },
    },
    {
      name: 'status',
      label: 'Status',
      options: {
        sort: true,
        sortThirdClickReset: true,
        customHeadRender: function CustomHead({ index, ...column }, sortColumn, columnMeta) {
          return (
            <SortableTableCell
              index={index}
              columnData={column}
              columnMeta={columnMeta}
              onSort={() => sortColumn(index)}
            />
          );
        },
        customBodyRender: function CustomBody(value, tableMeta) {
          const disabled = value === 'deleted' ? true : false;
          return (
            <>
              <FormControl>
                <Select
                  labelId="demo-simple-select-label"
                  id="demo-simple-select"
                  disabled={disabled}
                  value={value}
                  onChange={(e) => handleStatusChange(e, tableMeta.rowData[0])}
                  className={classes.statusSelect}
                  disableUnderline
                >
                  {statuses.map((s) => status(s))}
                </Select>
              </FormControl>
            </>
          );
        },
      },
    },
  ];

  const options = useMemo(
    () => ({
      filter: false,
      viewColumns: false,
      search: false,
      responsive: 'standard',
      resizableColumns: true,
      serverSide: true,
      count,
      rowsPerPage: pageSize,
      rowsPerPageOptions: [10, 20, 30],
      fixedHeader: true,
      page,
      print: false,
      download: false,
      textLabels: {
        selectedRows: {
          text: 'connection(s) selected',
        },
      },
      customToolbarSelect: (selected) => (
        <Button
          variant="contained"
          color="primary"
          size="large"
          // @ts-ignore
          onClick={() => handleDeleteConnections(selected)}
          style={{ background: '#8F1F00', marginRight: '10px' }}
        >
          <DeleteForeverIcon style={iconMedium} />
          Delete
        </Button>
      ),
      enableNestedDataAccess: '.',
      onTableChange: (action, tableState) => {
        const sortInfo = tableState.announceText ? tableState.announceText.split(' : ') : [];
        let order = '';
        if (tableState.activeColumn) {
          order = `${columns[tableState.activeColumn].name} desc`;
        }
        switch (action) {
          case 'changePage':
            setPage(tableState.page.toString());
            break;
          case 'changeRowsPerPage':
            setPageSize(tableState.rowsPerPage.toString());
            break;
          case 'sort':
            if (sortInfo.length == 2) {
              if (sortInfo[1] === 'ascending') {
                order = `${columns[tableState.activeColumn].name} asc`;
              } else {
                order = `${columns[tableState.activeColumn].name} desc`;
              }
            }
            if (order !== sortOrder) {
              setSortOrder(order);
            }
            break;
        }
      },
      expandableRows: true,
      expandableRowsHeader: false,
      expandableRowsOnClick: true,
      rowsExpanded: rowsExpanded,
      isRowExpandable: () => {
        return true;
      },
      onRowExpansionChange: (_, allRowsExpanded) => {
        setRowsExpanded(allRowsExpanded.slice(-1).map((item) => item.index));
        setShowMore(false);
      },
      renderExpandableRow: (rowData, tableMeta) => {
        const colSpan = rowData.length;
        const connection = connections && connections[tableMeta.rowIndex];

        return (
          <TableCell colSpan={colSpan} className={classes.innerTableWrapper}>
            <TableContainer className={classes.innerTableContainer}>
              <Table>
                <TableRow className={classes.noGutter}>
                  <TableCell style={{ padding: '20px 0' }}>
                    <Grid container spacing={1} style={{ textTransform: 'lowercase' }}>
                      <Grid item xs={12} md={12} className={classes.contentContainer}>
                        <Grid container spacing={1}>
                          <Grid
                            item
                            xs={12}
                            md={12}
                            style={{
                              display: 'flex',
                              flexWrap: 'wrap',
                              padding: '0 20px',
                              gap: 30,
                            }}
                            className={classes.contentContainer}
                          >
                            <FormattedMetadata event={connection} />
                          </Grid>
                        </Grid>
                      </Grid>
                    </Grid>
                  </TableCell>
                </TableRow>
              </Table>
            </TableContainer>
          </TableCell>
        );
      },
    }),
    [rowsExpanded, showMore, page, pageSize],
  );

  /**
   * fetch connections when the page loads
   */
  useEffect(() => {
    if (!loading) {
      getConnections(page, pageSize, search, sortOrder);
    }
  }, [page, pageSize, search, sortOrder]);

  const getConnections = (page, pageSize, search, sortOrder) => {
    setLoading(true);
    if (!search) search = '';
    if (!sortOrder) sortOrder = '';
    dataFetch(
      `/api/integrations/connections?page=${page}&pagesize=${pageSize}&search=${encodeURIComponent(
        search,
      )}&order=${encodeURIComponent(sortOrder)}`,
      {
        credentials: 'include',
        method: 'GET',
      },
      (res) => {
        setConnections(res?.connections || []);
        setPage(res?.page || 0);
        setCount(res?.total_count || 0);
        setPageSize(res?.page_size || 0);
        setLoading(false);
      },
      handleError(ACTION_TYPES.FETCH_CONNECTIONS),
    );
  };

  const handleError = (action) => (error) => {
    updateProgress({ showProgress: false });
    notify({
      message: `${action.error_msg}: ${error}`,
      event_type: EVENT_TYPES.ERROR,
      details: error.toString(),
    });
  };

  const handleStatusChange = (e, connectionId) => {
    const requestBody = JSON.stringify({
      status: e.target.value,
    });
    dataFetch(
      `/api/integrations/connections/${connectionId}`,
      {
        method: 'PUT',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: requestBody,
      },
      () => {
        getConnections(page, pageSize, search, sortOrder);
      },
      handleError(ACTION_TYPES.UPDATE_CONNECTION),
    );
  };

  const handleDeleteConnections = async (selected) => {
    if (selected) {
      let response = await modalRef.current.show({
        title: `Delete Connections`,
        subtitle: `Are you sure that you want to delete connections"?`,
        options: ['Delete', 'No'],
      });
      if (response === 'Delete') {
        selected.data.map(({ index }) => {
          deleteConnection(connections[index].id);
        });
      }
    }
  };

  const deleteConnection = (connectionId) => {
    dataFetch(
      `/api/integrations/connections/${connectionId}`,
      {
        method: 'DELETE',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
      },
      () => {
        getConnections(page, pageSize, search, sortOrder);
      },
      handleError(ACTION_TYPES.DELETE_CONNECTION),
    );
  };

  const [tableCols, updateCols] = useState(columns);

  const [columnVisibility, setColumnVisibility] = useState(() => {
    // Initialize column visibility based on the original columns' visibility
    const initialVisibility = {};
    columns.forEach((col) => {
      initialVisibility[col.name] = col.options?.display !== false;
    });
    return initialVisibility;
  });

  return (
    <>
      <NoSsr>
        <div className={StyleClass.toolWrapper}>
          <div className={classes.createButton}>
            <div>
              <Button
                aria-label="Rediscover"
                variant="contained"
                color="primary"
                size="large"
                // @ts-ignore
                onClick={onOpenCreateConnectionModal}
                style={{ marginRight: '2rem' }}
              >
                Connect Helm Repository
              </Button>
            </div>
          </div>
          <div
            className={classes.searchAndView}
            style={{
              display: 'flex',
            }}
          >
            {/* <Button
              aria-label="Edit"
              variant="contained"
              color="primary"
              size="large"
              // @ts-ignore
              onClick={() => {}}
              style={{ marginRight : "0.5rem" }}
            >
              <EditIcon style={iconMedium} />
            </Button> */}

            {/* <Button
              aria-label="Delete"
              variant="contained"
              color="primary"
              size="large"
              // @ts-ignore
              onClick={() => {}}
              style={{ background : "#8F1F00" }}
            >
              <DeleteForeverIcon style={iconMedium} />
              Delete
            </Button> */}

            <SearchBar
              onSearch={(value) => {
                setSearch(value);
              }}
              placeholder="Search connections..."
            />

            <CustomColumnVisibilityControl
              columns={columns}
              customToolsProps={{ columnVisibility, setColumnVisibility }}
            />
          </div>
        </div>
        {/* <div className={StyleClass.toolWrapper}>
          <div
            className={classes.bulkAction}
          >
            <Button
              aria-label="Delete"
              variant="contained"
              color="primary"
              size="large"
              // @ts-ignore
              onClick={() => {}}
              style={{ background : "#8F1F00" }}
            >
              <DeleteForeverIcon style={iconMedium} />
              Delete
            </Button>
          </div>
        </div> */}
        <ResponsiveDataTable
          data={connections}
          columns={columns}
          // @ts-ignore
          options={options}
          className={classes.muiRow}
          tableCols={tableCols}
          updateCols={updateCols}
          columnVisibility={columnVisibility}
        />
        <PromptComponent ref={modalRef} />
      </NoSsr>
    </>
  );
}

const SortableTableCell = ({ index, columnData, columnMeta, onSort }) => {
  return (
    <TableCell key={index} onClick={onSort}>
      <TableSortLabel
        active={columnMeta.name === columnData.name}
        direction={columnMeta.direction || 'asc'}
      >
        <b>{columnData.label}</b>
      </TableSortLabel>
    </TableCell>
  );
};

const mapDispatchToProps = (dispatch) => ({
  updateProgress: bindActionCreators(updateProgress, dispatch),
});

const mapStateToProps = (state) => {
  return {
    user: state.get('user')?.toObject(),
    selectedK8sContexts: state.get('selectedK8sContexts'),
  };
};

// @ts-ignore
export default withStyles(styles)(
  connect(mapStateToProps, mapDispatchToProps)(ConnectionManagementPage),
);
