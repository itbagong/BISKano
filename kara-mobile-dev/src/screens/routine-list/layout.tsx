/* eslint-disable @typescript-eslint/no-shadow */
/* eslint-disable react/no-unstable-nested-components */
/* eslint-disable react-native/no-inline-styles */
/* eslint-disable react-hooks/exhaustive-deps */
import {
  StyleSheet,
  Text,
  ScrollView,
  View,
  RefreshControl,
  ActivityIndicator,
  TouchableOpacity,
} from 'react-native';
import React from 'react';
import Pagination from '@components/pagination';
import container, {ContainerContext} from 'components/container';
import {useIsFocused} from '@react-navigation/native';
import {useActions} from '@overmind/index';
import {Colors, Mixins, Typography} from 'utils';
import s from '@components/styles/index';
import {Button, Icon} from '@ui-kitten/components/ui';
import AccordionItem from '@components/accordion-item';
import ModalFilter from './modal-filter';
import ModalAddNew from './modal-add-new';
import moment from 'moment';
import {ALERT_TYPE, Dialog, Toast} from 'react-native-alert-notification';

import {default as theme} from '../../custom-theme.json';

type Props = {
  navigation: any;
};

const ListScreen = (props: Props) => {
  const {navigation} = props;
  const isFocused = useIsFocused();
  const ctx = React.useContext(ContainerContext);
  const {getsRoutine, deleteRoutine, getRBAC, getsSite} = useActions();
  const [loading, setLoading] = React.useState(false);
  const [data, setData] = React.useState({} as any);
  const [activePage, setActivePage] = React.useState(1);
  const [openfilter, setOpenFilter] = React.useState(false);
  const [openModalAddNew, setOpenModalAddNew] = React.useState(false);
  const [fprofile, setFprofile] = React.useState({} as any);

  const [filter, setFilter] = React.useState({
    Skip: 0,
    Take: 10,
    Sort: ['-_id'],
  } as any);
  const resetFilter = () => {
    setFilter({
      Skip: 0,
      Take: 10,
      Sort: ['-ExecutionDate'],
    });
  };
  const initData = () => {
    if (fprofile?.canRead) {
      setLoading(true);
      getsRoutine(filter)
        .then((res: any) => {
          setData(res);
        })
        .finally(() => setLoading(false));
    }
  };

  React.useLayoutEffect(() => {
    navigation.setOptions({
      headerRight: () => (
        <TouchableOpacity
          style={styles.buttonFilter}
          onPress={() => setOpenFilter(true)}>
          <Icon
            style={{
              width: Mixins.scaleSize(20),
              height: Mixins.scaleSize(20),
            }}
            fill="#ffffff"
            name="funnel-outline"
          />
          <Text style={styles.labelButtonFilter}>Filter</Text>
        </TouchableOpacity>
      ),
    });
    ctx.setRefreshCallback({
      func: async () => {
        initData();
      },
    });
    return () => {};
  }, [isFocused]);
  React.useEffect(() => {
    getRBAC('Routine').then(r => {
      setFprofile(r);
    });
    setActivePage(1);
    resetFilter();
    return () => {};
  }, [isFocused]);

  React.useEffect(() => {
    if (isFocused) {
      if (filter?.Where) {
        initData();
      } else {
        getsSite({Select: ['_id', 'Name'], Sort: ['_id']}).then(res => {
          if (res.length === 1) {
            applyFilter({site: res[0]._id});
          } else {
            initData();
          }
        });
      }
    }
    return () => {};
  }, [isFocused, filter, fprofile]);
  const applyFilter = (_filter: any) => {
    let queries = [];
    if (_filter?.keyword) {
      const query = {
        Op: '$or',
        Items: [
          {
            Field: '_id',
            Op: '$contains',
            Value: [_filter?.keyword],
          },
          {
            Field: 'Name',
            Op: '$contains',
            Value: [_filter?.keyword],
          },
        ],
      };
      queries.push(query);
    }
    if (_filter?.site) {
      queries.push({
        Op: '$in',
        Field: 'SiteID',
        Value: [_filter?.site],
      });
    }
    if (_filter?.executionDate?.from) {
      queries.push({
        Op: '$gte',
        Field: 'ExecutionDate',
        Value: _filter?.executionDate?.from,
      });
    }
    if (_filter?.executionDate?.to) {
      queries.push({
        Op: '$lte',
        Field: 'ExecutionDate',
        Value: _filter?.executionDate?.to,
      });
    }
    // let newItems = filter.Where.Items.map((item: any) => {
    //   item.Value = [keyword];
    //   return item;
    // });
    if (queries.length === 1) {
      setFilter({...filter, Where: {...queries[0]}});
    } else if (queries.length > 1) {
      setFilter({...filter, Where: {Op: '$and', Items: queries}});
    }
    setOpenFilter(false);
  };

  const onDelete = (item: any) => {
    Dialog.show({
      type: ALERT_TYPE.DANGER,
      title: 'Delete',
      textBody:
        'You will delete data ! Are you sure ? \n Please be noted, this can not be undone !',
      button: 'Submit',
      onPressButton: () => {
        deleteRoutine(item)
          .then(() => {
            Toast.show({
              type: ALERT_TYPE.SUCCESS,
              title: 'Success',
              textBody: 'Data has been deleted!',
            });
            initData();
          })
          .finally(() => {
            Dialog.hide();
          });
      },
    });
  };
  return (
    <>
      {isFocused && (
        <View style={s.container}>
          <View style={{flex: 1}}>
            {loading ? (
              <View
                style={{
                  flex: 1,
                  justifyContent: 'center',
                  alignItems: 'center',
                }}>
                <ActivityIndicator color={Colors.PRIMARY.red} />
              </View>
            ) : (
              <ScrollView
                contentInsetAdjustmentBehavior="automatic"
                refreshControl={
                  <RefreshControl refreshing={loading} onRefresh={initData} />
                }>
                {data?.data?.map((item: any, i: number) => (
                  <AccordionItem key={i} title={item._id}>
                    <View style={styles.rowForm}>
                      <Text
                        style={{
                          flex: 1,
                          ...Typography.textMdPlus,
                          color: Colors.BLACK,
                        }}>
                        Site
                      </Text>
                      <Text
                        style={{
                          flex: 1,
                          ...Typography.textMdPlusSemiBold,
                          color: Colors.BLACK,
                        }}>
                        : {item.SiteName}
                      </Text>
                    </View>
                    <View style={styles.rowForm}>
                      <Text
                        style={{
                          flex: 1,
                          ...Typography.textMdPlus,
                          color: Colors.BLACK,
                        }}>
                        Execution date
                      </Text>
                      <Text
                        style={{
                          flex: 1,
                          ...Typography.textMdPlusSemiBold,
                          color: Colors.BLACK,
                        }}>
                        : {moment(item.ExecutionDate).format('DD MMM YYYY')}
                      </Text>
                    </View>
                    <View
                      style={{
                        ...styles.rowForm,
                        gap: Mixins.scaleSize(10),
                        marginTop: Mixins.scaleSize(10),
                      }}>
                      {fprofile?.canUpdate ? (
                        <TouchableOpacity
                          onPress={() => {
                            navigation.navigate('RoutineDetails', {
                              _id: item._id,
                            });
                          }}
                          style={{
                            ...styles.buttonAction,
                            backgroundColor: theme['color-success-300'],
                          }}>
                          <Icon
                            fill={theme['color-success-900']}
                            name="edit-2-outline"
                            style={{
                              width: Mixins.scaleSize(20),
                              height: Mixins.scaleSize(20),
                            }}
                          />
                          <Text
                            style={{
                              ...styles.labelAction,
                              color: theme['color-success-900'],
                            }}>
                            Edit
                          </Text>
                        </TouchableOpacity>
                      ) : (
                        <>
                          {fprofile?.canRead && (
                            <TouchableOpacity
                              onPress={() => {
                                navigation.navigate('RoutineDetails', {
                                  _id: item._id,
                                  isReadOnly: true,
                                });
                              }}
                              style={{
                                ...styles.buttonAction,
                                backgroundColor: theme['color-success-300'],
                              }}>
                              <Icon
                                fill={theme['color-success-900']}
                                name="eye-outline"
                                style={{
                                  width: Mixins.scaleSize(20),
                                  height: Mixins.scaleSize(20),
                                }}
                              />
                              <Text
                                style={{
                                  ...styles.labelAction,
                                  color: theme['color-success-900'],
                                }}>
                                View
                              </Text>
                            </TouchableOpacity>
                          )}
                        </>
                      )}
                      {fprofile?.canDelete && (
                        <TouchableOpacity
                          onPress={() => onDelete(item)}
                          style={{
                            ...styles.buttonAction,
                            backgroundColor: theme['color-primary-300'],
                          }}>
                          <Icon
                            fill={theme['color-primary-900']}
                            name="trash-2-outline"
                            style={{
                              width: Mixins.scaleSize(20),
                              height: Mixins.scaleSize(20),
                            }}
                          />
                          <Text
                            style={{
                              ...styles.labelAction,
                              color: theme['color-primary-900'],
                            }}>
                            Delete
                          </Text>
                        </TouchableOpacity>
                      )}
                    </View>
                  </AccordionItem>
                ))}
              </ScrollView>
            )}
          </View>

          {data?.count > 0 && !loading && (
            <Pagination
              dataCount={data.count}
              showPer={filter.Take}
              state={activePage}
              containerStyle={{marginVertical: Mixins.scaleSize(10)}}
              onChangeState={(value: number) => {
                setActivePage(value);
                setFilter({...filter, Skip: (value - 1) * filter.Take});
              }}
            />
          )}
          {fprofile?.canCreate && (
            <Button
              onPress={() => setOpenModalAddNew(true)}
              style={styles.buttonAdd}
              size="large"
              status="primary"
              accessoryLeft={() => (
                <Icon
                  style={{
                    width: Mixins.scaleSize(22),
                    height: Mixins.scaleSize(22),
                  }}
                  fill="#ffffff"
                  name="plus-outline"
                />
              )}>
              {() => (
                <Text
                  style={{
                    ...Typography.textLgSemiBold,
                    color: 'white',
                    marginLeft: Mixins.scaleSize(10),
                  }}>
                  Add new
                </Text>
              )}
            </Button>
          )}
          <ModalFilter
            isOpen={openfilter}
            onClose={() => setOpenFilter(false)}
            onReset={() => {
              setActivePage(1);
              resetFilter();
            }}
            onApply={applyFilter}
          />
          <ModalAddNew
            isOpen={openModalAddNew}
            onClose={() => setOpenModalAddNew(false)}
            navigation={navigation}
          />
        </View>
      )}
    </>
  );
};

export default container(ListScreen, false);

const styles = StyleSheet.create({
  menucontent: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: Mixins.scaleSize(5),
    paddingVertical: Mixins.scaleSize(10),
    backgroundColor: Colors.WHITE,
  },
  buttonAdd: {
    borderRadius: Mixins.scaleSize(8),
    alignItems: 'center',
    justifyContent: 'center',
  },
  rowForm: {
    flexDirection: 'row',
  },
  buttonAction: {
    flex: 1,
    flexDirection: 'row',
    padding: Mixins.scaleSize(10),
    borderRadius: Mixins.scaleSize(5),
    justifyContent: 'center',
    alignItems: 'center',
    gap: Mixins.scaleSize(10),
  },
  labelAction: {
    ...Typography.textMdPlusSemiBold,
  },
  buttonFilter: {
    flexDirection: 'row',
    backgroundColor: Colors.SHADES.gray[600],
    marginRight: Mixins.scaleSize(14),
    padding: Mixins.scaleSize(7),
    paddingHorizontal: Mixins.scaleSize(10),
    gap: Mixins.scaleSize(10),
    borderRadius: Mixins.scaleSize(5),
  },
  labelButtonFilter: {
    ...Typography.textMdPlus,
    color: Colors.WHITE,
  },
});
