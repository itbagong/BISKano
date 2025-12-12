/* eslint-disable react-hooks/exhaustive-deps */
/* eslint-disable react-native/no-inline-styles */
import {
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import React from 'react';
import s from '@components/styles/index';
import {Mixins, Colors, Typography} from 'utils';
import moment from 'moment';
import {Card, Datepicker, Icon, Input} from '@ui-kitten/components/ui';
import {useAppState, useActions} from '@overmind/index';
import Dropdown from 'components/dropdown';
import {default as theme} from '../../custom-theme.json';
import DateTimePickerModal from 'react-native-modal-datetime-picker';
import TimePicker from '@components/timepicker';
type Props = {
  routineAsset: any;
  routine: any;
  navigation: any;
  isReadOnly: any;
};

const TabChecklist = (props: Props) => {
  const {routineAsset, routine, navigation} = props;
  const {routineChecklist} = useAppState();
  const {
    findEmployee,
    findDepartment,
    findMasterData,
    findSite,
    changeDataRoutineChecklist,
  } = useActions();
  // const [status, setStatus] = React.useState(routineAsset?.StatusCondition);
  const [masterEmployees, setMasterEmployees] = React.useState([]);
  const [masterEmployeesForMarketing, setmasterEmployeesForMarketing] =
    React.useState([]);
  const [masterDepartment, setMasterDepartment] = React.useState([]);
  const [masterShift, setMasterShift] = React.useState([]);
  const [masterWorkLoc, setMasterWorkLoc] = React.useState([]);
  const [showTimeBreakdown, setShowTimeBreakdown] = React.useState(false);
  const getLabelStatus = (_status: string) => {
    switch (_status) {
      case 'NotCheckedYet':
        return 'Not Checked Yet';
      case 'NeedRepair':
        return 'Need Repair';
      case 'RunningWell':
        return 'Running Well';
      default:
        break;
    }
    return _status;
  };
  const getEmployees = () => {
    return findEmployee({
      Select: ['_id', 'Name'],
      Sort: ['Name'],
      Take: 20,
      Where: {
        Items: [
          {
            Field: 'Name',
            Op: '$contains',
            Value: [routineChecklist?.RoutineChecklist?.Name],
          },
          {
            Field: '_id',
            Op: '$contains',
            Value: [routineChecklist?.RoutineChecklist?.Name],
          },
        ],
        Op: '$or',
      },
    });
  };
  const getMarketing = () => {
    return findEmployee({
      Select: ['_id', 'Name'],
      Sort: ['Name'],
      Take: 20,
      Where: {
        Items: [
          {
            Field: 'Name',
            Op: '$contains',
            Value: [routineChecklist?.RoutineChecklist?.Marketing],
          },
          {
            Field: '_id',
            Op: '$contains',
            Value: [routineChecklist?.RoutineChecklist?.Marketing],
          },
        ],
        Op: '$or',
      },
    });
  };
  const getDepartment = () => {
    return findDepartment({
      Select: ['_id', 'Label'],
      Sort: ['Label'],
    });
  };
  const getShift = () => {
    return findMasterData({
      Select: ['_id', 'Name'],
      Sort: ['Name'],
      MasterDataTypeID: 'SHFT',
    });
  };
  const getWorkLocation = () => {
    return findSite({
      Select: ['_id', 'Label'],
      Sort: ['Label'],
      Where: {
        Field: '_id',
        Op: '$contains',
        Value: [routineChecklist?.RoutineChecklist?.WorkLocation],
      },
    });
  };
  const init = () => {
    Promise.all([
      getEmployees(),
      getMarketing(),
      getDepartment(),
      getShift(),
      getWorkLocation(),
    ]).then(res => {
      const [emps, marketings, departs, shifts, workLocs] = res;
      setMasterEmployees(emps);
      setmasterEmployeesForMarketing(marketings);
      setMasterDepartment(departs);
      setMasterShift(shifts);
      setMasterWorkLoc(workLocs);
    });
  };
  React.useEffect(() => {
    init();

    return () => {};
  }, []);
  // React.useEffect(() => {
  //   return () => {};
  // }, [routineChecklist]);

  const setValueDateTime = () => {
    return new Date(
      moment().format('YYYY-MM-DD') +
        ' ' +
        routineChecklist?.RoutineChecklist?.TimeBreakdown,
    );
  };
  const onSearchMarketing = (text: string) => {
    findEmployee({
      Select: ['_id', 'Name'],
      Sort: ['Name'],
      Take: 20,
      Where: {
        Items: [
          {
            Field: 'Name',
            Op: '$contains',
            Value: [text],
          },
          {
            Field: '_id',
            Op: '$contains',
            Value: [text],
          },
        ],
        Op: '$or',
      },
    }).then(res => {
      setmasterEmployeesForMarketing(res);
    });
  };
  return (
    <View
      style={{
        ...s.container,
        backgroundColor: Colors.SHADES.gray[100],
      }}>
      <View style={{flex: 1, backgroundColor: Colors.SHADES.gray[100]}}>
        <ScrollView>
          <View>
            <Card style={styles.card}>
              <View
                style={{
                  ...s.row,
                  flex: 1,
                  justifyContent: 'space-between',
                  alignItems: 'center',
                }}>
                <View style={{...s.column}}>
                  <Text style={{...Typography.textMd, color: Colors.BLACK}}>
                    {routineAsset?._id} | {routineAsset?.AssetName}
                  </Text>
                  <Text style={{...Typography.textMd, color: Colors.BLACK}}>
                    {moment(routine?.ExecutionDate).format('DD MMM YYYY')}
                  </Text>
                </View>
                <Text
                  style={{...Typography.textMdSemiBold, color: Colors.BLACK}}>
                  {getLabelStatus(routineAsset?.StatusCondition)}
                </Text>
              </View>
            </Card>
            <View style={styles.form}>
              <Dropdown
                items={masterEmployees.map((o: any) => {
                  return {value: o._id, label: o.Name};
                })}
                disabled
                value={routineChecklist?.RoutineChecklist?.Name}
                onSelect={(selected: any) => {
                  changeDataRoutineChecklist({
                    ...routineChecklist,
                    RoutineChecklist: {
                      ...routineChecklist.RoutineChecklist,
                      Name: selected,
                    },
                  });
                }}
                inputLabel="Name"
              />
              <Dropdown
                items={masterDepartment.map((o: any) => {
                  return {value: o._id, label: o.Label};
                })}
                disabled={
                  routineChecklist?.RoutineChecklist?.IsAlreadyRequest ||
                  props.isReadOnly
                }
                value={routineChecklist?.RoutineChecklist?.Department}
                onSelect={(selected: any) => {
                  changeDataRoutineChecklist({
                    ...routineChecklist,
                    RoutineChecklist: {
                      ...routineChecklist.RoutineChecklist,
                      Department: selected,
                    },
                  });
                }}
                inputLabel="Department"
              />
              <Dropdown
                items={masterShift.map((o: any) => {
                  return {value: o._id, label: o.Name};
                })}
                disabled={
                  routineChecklist?.RoutineChecklist?.IsAlreadyRequest ||
                  props.isReadOnly
                }
                value={routineChecklist?.RoutineChecklist?.Shift}
                onSelect={(selected: any) => {
                  changeDataRoutineChecklist({
                    ...routineChecklist,
                    RoutineChecklist: {
                      ...routineChecklist.RoutineChecklist,
                      Shift: selected,
                    },
                  });
                }}
                inputLabel="Shift"
              />
              <Dropdown
                items={masterWorkLoc.map((o: any) => {
                  return {value: o._id, label: o.Label};
                })}
                // disabled
                disabled={
                  routineChecklist?.RoutineChecklist?.IsAlreadyRequest ||
                  props.isReadOnly
                }
                value={routineChecklist?.RoutineChecklist?.WorkLocation}
                onSelect={(selected: any) => {
                  changeDataRoutineChecklist({
                    ...routineChecklist,
                    RoutineChecklist: {
                      ...routineChecklist.RoutineChecklist,
                      WorkLocation: selected,
                    },
                  });
                }}
                inputLabel="Work Location"
              />
              <View style={{marginBottom: Mixins.scaleSize(10)}}>
                <Text style={[styles.label]}>KM Today</Text>
                <Input
                  disabled={
                    routineChecklist?.RoutineChecklist?.IsAlreadyRequest ||
                    props.isReadOnly
                  }
                  value={routineChecklist?.RoutineChecklist?.KmToday.toString()}
                  style={{textAlign: 'right'}}
                  textStyle={{
                    paddingHorizontal: 0,
                    textAlign: 'right',
                    color: Colors.BLACK,
                  }}
                  placeholder=""
                  keyboardType="number-pad"
                  onChangeText={nextValue => {
                    changeDataRoutineChecklist({
                      ...routineChecklist,
                      RoutineChecklist: {
                        ...routineChecklist.RoutineChecklist,
                        KmToday: nextValue,
                      },
                    });
                  }}
                />
              </View>
              <View style={{marginBottom: Mixins.scaleSize(10)}}>
                <Text style={[styles.label]}>Time breakdown</Text>
                <TouchableOpacity
                  style={{
                    flex: 1,
                    backgroundColor: '#f7f9fc',
                    borderColor: '#e4e9f2',
                    borderWidth: 1,
                    borderRadius: 4,
                    paddingVertical: 8,
                    paddingHorizontal: 14,
                  }}
                  disabled={
                    routineChecklist?.RoutineChecklist?.IsAlreadyRequest ||
                    props.isReadOnly
                  }
                  onPress={() => setShowTimeBreakdown(true)}>
                  <Text style={{...Typography.textMdPlus, color: Colors.BLACK}}>
                    {routineChecklist?.RoutineChecklist?.TimeBreakdown
                      ? routineChecklist?.RoutineChecklist?.TimeBreakdown
                      : '--:--'}
                  </Text>
                </TouchableOpacity>
                <DateTimePickerModal
                  isVisible={showTimeBreakdown}
                  date={
                    routineChecklist?.RoutineChecklist?.TimeBreakdown
                      ? setValueDateTime()
                      : new Date()
                  }
                  style={{borderRadius: Mixins.scaleSize(10)}}
                  mode="time"
                  is24Hour
                  onConfirm={date => {
                    changeDataRoutineChecklist({
                      ...routineChecklist,
                      RoutineChecklist: {
                        ...routineChecklist.RoutineChecklist,
                        TimeBreakdown: moment(date).format('HH:mm'),
                      },
                    });
                    setShowTimeBreakdown(false);
                  }}
                  onCancel={() => setShowTimeBreakdown(false)}
                />
              </View>
              <View
                style={{
                  marginBottom: Mixins.scaleSize(10),
                  flexDirection: 'row',
                  gap: 5,
                }}>
                <View style={{flex: 1.5}}>
                  <Text style={[styles.label]}>Departure date</Text>
                  <Datepicker
                    date={
                      routineChecklist?.RoutineChecklist?.Departure
                        ? new Date(
                            routineChecklist?.RoutineChecklist?.Departure,
                          )
                        : new Date()
                    }
                    disabled={
                      routineChecklist?.RoutineChecklist?.IsAlreadyRequest ||
                      props.isReadOnly
                    }
                    onSelect={nextDate => {
                      const time = moment(
                        routineChecklist?.RoutineChecklist?.Departure,
                      ).format('HH:mm');

                      changeDataRoutineChecklist({
                        ...routineChecklist,
                        RoutineChecklist: {
                          ...routineChecklist.RoutineChecklist,
                          Departure: moment(
                            moment(nextDate).format('DD-MM-YYYY') + ' ' + time,
                            'DD-MM-YYYY HH:mm',
                          ),
                        },
                      });
                    }}
                  />
                </View>
                <View style={{flex: 1}}>
                  <TimePicker
                    label="Departure time"
                    disabled={
                      routineChecklist?.RoutineChecklist?.IsAlreadyRequest ||
                      props.isReadOnly
                    }
                    defaultDate={
                      routineChecklist?.RoutineChecklist?.Departure
                        ? routineChecklist?.RoutineChecklist?.Departure
                        : new Date()
                    }
                    value={moment(
                      routineChecklist?.RoutineChecklist?.Departure
                        ? routineChecklist?.RoutineChecklist?.Departure
                        : new Date(),
                    ).format('HH:mm')}
                    onChange={date =>
                      changeDataRoutineChecklist({
                        ...routineChecklist,
                        RoutineChecklist: {
                          ...routineChecklist.RoutineChecklist,
                          Departure: date,
                        },
                      })
                    }
                  />
                </View>
              </View>
              <View
                style={{
                  marginBottom: Mixins.scaleSize(10),
                  flexDirection: 'row',
                  gap: 5,
                }}>
                <View style={{flex: 1.5}}>
                  <Text style={[styles.label]}>Arrival date</Text>
                  <Datepicker
                    date={
                      routineChecklist?.RoutineChecklist?.Arrive
                        ? new Date(routineChecklist?.RoutineChecklist?.Arrive)
                        : new Date()
                    }
                    disabled={
                      routineChecklist?.RoutineChecklist?.IsAlreadyRequest ||
                      props.isReadOnly
                    }
                    onSelect={nextDate => {
                      const time = moment(
                        routineChecklist?.RoutineChecklist?.Arrive,
                      ).format('HH:mm');

                      changeDataRoutineChecklist({
                        ...routineChecklist,
                        RoutineChecklist: {
                          ...routineChecklist.RoutineChecklist,
                          Arrive: moment(
                            moment(nextDate).format('DD-MM-YYYY') + ' ' + time,
                            'DD-MM-YYYY HH:mm',
                          ),
                        },
                      });
                    }}
                  />
                </View>
                <View style={{flex: 1}}>
                  <TimePicker
                    label="Arrival time"
                    disabled={
                      routineChecklist?.RoutineChecklist?.IsAlreadyRequest ||
                      props.isReadOnly
                    }
                    defaultDate={
                      routineChecklist?.RoutineChecklist?.Arrive
                        ? routineChecklist?.RoutineChecklist?.Arrive
                        : new Date()
                    }
                    value={moment(
                      routineChecklist?.RoutineChecklist?.Arrive
                        ? routineChecklist?.RoutineChecklist?.Arrive
                        : new Date(),
                    ).format('HH:mm')}
                    onChange={date =>
                      changeDataRoutineChecklist({
                        ...routineChecklist,
                        RoutineChecklist: {
                          ...routineChecklist.RoutineChecklist,
                          Arrive: date,
                        },
                      })
                    }
                  />
                </View>
              </View>
              <View style={{marginBottom: Mixins.scaleSize(10)}}>
                <Text style={[styles.label]}>BBM Level</Text>
                <Input
                  disabled={
                    routineChecklist?.RoutineChecklist?.IsAlreadyRequest ||
                    props.isReadOnly
                  }
                  value={routineChecklist?.RoutineChecklist?.BBMLevel}
                  textStyle={{paddingHorizontal: 0}}
                  placeholder=""
                  onChangeText={nextValue => {
                    changeDataRoutineChecklist({
                      ...routineChecklist,
                      RoutineChecklist: {
                        ...routineChecklist.RoutineChecklist,
                        BBMLevel: nextValue,
                      },
                    });
                  }}
                />
              </View>
              <View style={{marginBottom: Mixins.scaleSize(10)}}>
                <Text style={[styles.label]}>Helper</Text>
                <Input
                  disabled={
                    routineChecklist?.RoutineChecklist?.IsAlreadyRequest ||
                    props.isReadOnly
                  }
                  value={routineChecklist?.RoutineChecklist?.HelperName}
                  textStyle={{paddingHorizontal: 0}}
                  placeholder=""
                  onChangeText={nextValue => {
                    changeDataRoutineChecklist({
                      ...routineChecklist,
                      RoutineChecklist: {
                        ...routineChecklist.RoutineChecklist,
                        HelperName: nextValue,
                      },
                    });
                  }}
                />
              </View>
              <View style={{marginBottom: Mixins.scaleSize(10)}}>
                <Text style={[styles.label]}>Driver 1</Text>
                <Input
                  disabled={
                    routineChecklist?.RoutineChecklist?.IsAlreadyRequest ||
                    props.isReadOnly
                  }
                  value={routineChecklist?.RoutineChecklist?.Driver1}
                  textStyle={{paddingHorizontal: 0}}
                  placeholder=""
                  onChangeText={nextValue => {
                    changeDataRoutineChecklist({
                      ...routineChecklist,
                      RoutineChecklist: {
                        ...routineChecklist.RoutineChecklist,
                        Driver1: nextValue,
                      },
                    });
                  }}
                />
              </View>
              <View style={{marginBottom: Mixins.scaleSize(10)}}>
                <Text style={[styles.label]}>Driver 2</Text>
                <Input
                  disabled={
                    routineChecklist?.RoutineChecklist?.IsAlreadyRequest ||
                    props.isReadOnly
                  }
                  value={routineChecklist?.RoutineChecklist?.Driver2}
                  textStyle={{paddingHorizontal: 0}}
                  placeholder=""
                  onChangeText={nextValue => {
                    changeDataRoutineChecklist({
                      ...routineChecklist,
                      RoutineChecklist: {
                        ...routineChecklist.RoutineChecklist,
                        Driver2: nextValue,
                      },
                    });
                  }}
                />
              </View>
              <Dropdown
                items={masterEmployeesForMarketing.map((o: any) => {
                  return {value: o._id, label: o.Name};
                })}
                disabled={
                  routineChecklist?.RoutineChecklist?.IsAlreadyRequest ||
                  props.isReadOnly
                }
                value={routineChecklist?.RoutineChecklist?.Marketing}
                onSelect={(selected: any) => {
                  changeDataRoutineChecklist({
                    ...routineChecklist,
                    RoutineChecklist: {
                      ...routineChecklist.RoutineChecklist,
                      Marketing: selected,
                    },
                  });
                }}
                inputLabel="Marketing"
                onSearch={onSearchMarketing}
              />
              <View style={{marginBottom: Mixins.scaleSize(10)}}>
                <ScrollView nestedScrollEnabled horizontal>
                  <View
                    style={{flexDirection: 'row', gap: Mixins.scaleSize(10)}}>
                    {routineChecklist?.RoutineChecklistCategories?.map(
                      (item: any, i: number) => {
                        const len =
                          routineChecklist.RoutineChecklistCategories.length >=
                          3
                            ? 3
                            : routineChecklist.RoutineChecklistCategories
                                .length;
                        return (
                          <TouchableOpacity
                            key={i}
                            style={{
                              ...styles.boxCategory,
                              width:
                                (Mixins.WINDOW_WIDTH - Mixins.scaleSize(48)) /
                                len,
                            }}
                            onPress={() => {
                              navigation.navigate(
                                'RoutineDetailsAssetChecklist',
                                {
                                  headerTitle: item.CategoryName,
                                  CategoryID: item.CategoryID,
                                  index: i,
                                  isReadOnly: props.isReadOnly,
                                },
                              );
                            }}>
                            <Icon
                              fill={theme['color-primary-500']}
                              name="cube-outline"
                              style={{
                                width: Mixins.scaleSize(30),
                                height: Mixins.scaleSize(30),
                              }}
                            />
                            <Text
                              style={{
                                ...Typography.textMdSemiBold,
                                color: Colors.BLACK,
                                textAlign: 'center',
                              }}>
                              {item.CategoryName}
                            </Text>
                          </TouchableOpacity>
                        );
                      },
                    )}
                  </View>
                </ScrollView>
              </View>
            </View>
          </View>
        </ScrollView>
      </View>
    </View>
  );
};

export default TabChecklist;

const styles = StyleSheet.create({
  card: {
    borderRadius: Mixins.scaleSize(12),
    marginBottom: Mixins.scaleSize(15),
  },
  rowForm: {
    flexDirection: 'row',
  },
  form: {},
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
  label: {
    ...Typography.textMdPlus,
    color: Colors.BLACK,
  },
  boxCategory: {
    height: Mixins.scaleSize(100),
    // borderWidth: 1,
    borderRadius: Mixins.scaleSize(10),
    // borderColor: theme['color-primary-500'],
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: 'white',
  },
});
